# Audio Chunker

The `chunker` package turns a continuous stream of audio frames into bounded
audio chunks that can be sent to Whisper. It uses an RMS energy threshold to
separate speech from silence, emits replaceable partial chunks while speech is
in progress, and emits a final chunk when the utterance ends or becomes too
long.

The package does not perform linguistic voice activity detection (VAD). A
frame whose energy exceeds the threshold is considered speech, so sufficiently
loud music, clicks, or background noise can also start an utterance.

## Processing flow

```text
FFmpeg PCM stream
    -> 100 ms frame
    -> RMS speech/silence classification
    -> utterance buffer with pre-roll and trailing padding
    -> partial or final AudioChunk
    -> Whisper transcription
```

Call `AddFrame` for each frame in chronological order. Call `Flush` when the
stream stops so that a pending utterance is not lost.

## Audio terms and units

| Term | Meaning in this package |
| --- | --- |
| **PCM** | Pulse-code modulation: audio represented as a sequence of amplitude values rather than a compressed format. FFmpeg supplies headerless, mono, little-endian `float32` PCM (`f32le`). |
| **Sample** | One `float32` amplitude measurement at one point in time. Samples are normalized PCM values, normally in the range `-1.0` to `1.0`. |
| **Sample rate** | Number of samples per second. The default is 16,000 Hz, so 16,000 samples represent one second of audio. |
| **Sample offset** | A zero-based position in the complete input stream, measured in samples. The chunker uses offsets internally to calculate timestamps without accumulating frame-duration rounding errors. |
| **Frame** | One slice of consecutive input samples passed to `AddFrame`. FFmpeg normally produces 100 ms frames, or 1,600 samples at 16 kHz. Classification applies to the entire frame. |
| **Amplitude** | The instantaneous magnitude and sign of a sample. The chunker squares amplitudes while calculating frame energy, so positive and negative values contribute equally. |
| **RMS amplitude** | Root mean square amplitude, used as a simple measure of a frame's signal energy: `sqrt(sum(sample^2) / sampleCount)`. |
| **Energy threshold** | The minimum RMS amplitude for a frame to be classified as speech. The default is `0.01`. |
| **Speech frame** | A non-empty frame whose RMS amplitude is greater than or equal to `energyThreshold`. |
| **Silence frame / non-speech frame** | A frame whose RMS amplitude is below `energyThreshold`. “Silence” therefore means low energy, not necessarily digital zero. |
| **Duration** | Elapsed audio time. Conversion uses `sampleCount / sampleRate`; at 16 kHz, one sample is 62.5 microseconds. |

The chunker assumes every frame uses the configured sample rate and arrives
without gaps, reordering, or duplicated samples. It does not resample audio or
validate frame length.

## Speech and chunk terms

| Term | Meaning in this package |
| --- | --- |
| **Utterance** | The buffered region that starts when a speech frame is detected and ends after sustained silence, an enforced duration limit, or `Flush`. It can contain speech plus retained silence. |
| **Active speech** | Samples belonging to frames classified as speech. Only these samples count toward `minSpeech`; pre-roll and silence frames do not. |
| **Pre-roll** | Recent audio retained while idle and prepended when speech starts. It preserves the beginning of words when frame-level detection reacts late. Its normal limit is `speechPad`. |
| **Speech padding** | Low-energy context retained before and after detected speech. The default is 300 ms. |
| **Trailing silence** | The current uninterrupted run of non-speech samples after the last speech frame. Once it reaches `silenceToFinal`, all but `speechPad` of it is removed from the final chunk. |
| **Partial chunk** | A non-final, provisional view of the current utterance. It contains at most the latest `partialWindow` of buffered audio and may overlap or replace earlier partial results. `Final` is `false`. |
| **Final chunk** | A completed transcription unit. It normally contains the whole buffered utterance, including padding, and has `Final` set to `true`. |
| **Forced final** | A final chunk emitted when the buffer reaches `maxFinalDuration`, even if speech has not stopped. This bounds transcription work and latency. |
| **Overlap** | Audio copied from the end of a forced final and prepended to the next utterance if speech continues. It gives Whisper context across the split. Consecutive final chunks can therefore cover some of the same time. |
| **Flush** | Explicit completion of a pending utterance when input ends or recording is cancelled. It does not wait for trailing silence. |
| **Timestamp** | A chunk's `Start` or `End` offset from the beginning of the input stream. Timestamps describe the included sample range, not wall-clock time. |

## Public API

### `AudioChunk`

An `AudioChunk` is an owned copy of audio ready for transcription.

| Field | Meaning |
| --- | --- |
| `Samples` | The normalized PCM samples included in the chunk. The slice is copied so later chunker mutations do not change an emitted chunk. |
| `Start` | Inclusive offset of the first included sample from the start of the stream. A sliding partial can start later than its utterance, and an overlapped final can start before the preceding final ended. |
| `End` | Exclusive offset immediately after the last included sample. For a chunk, `End - Start` equals the duration represented by `Samples`, subject to `time.Duration` conversion precision. |
| `Final` | `false` for a provisional partial and `true` for a completed chunk. |

### `AudioChunker`

`AudioChunker` is the stateful stream processor. One instance represents one
ordered audio stream and is not safe for concurrent calls.

| Operation | Meaning |
| --- | --- |
| `NewAudioChunker()` | Creates an idle chunker using a copy of `DefaultConfig`. |
| `AddFrame(samples)` | Advances the stream by the number of supplied samples, classifies the frame, updates the current utterance, and returns zero or more newly emitted chunks. An empty frame is ignored. |
| `Flush()` | Emits one final chunk for a pending utterance if it contains at least `minSpeech`, then resets utterance state. It returns nothing when idle or when the buffered speech is too short. |

`AddFrame` returns a slice because one speech frame can cross both the partial
interval and maximum-duration boundaries. In that case it emits a partial
chunk followed by a final chunk.

## Configuration terms

All `Config` fields are currently package-private. The table documents the
behavior controlled by `DefaultConfig` and supports maintenance inside this
package; callers outside `chunker` cannot set individual fields directly.

| Field | Default | Meaning |
| --- | ---: | --- |
| `sampleRate` | 16,000 Hz | Input samples per second and the basis of every sample/duration conversion. It must match the recorder output. |
| `frameDuration` | 100 ms | Expected input-frame duration. The chunking algorithm does not enforce it; it is also used by package tests to construct frames. |
| `minSpeech` | 250 ms | Minimum cumulative duration of speech-classified frames required before any partial or final chunk is valid. Short bursts are discarded as noise. |
| `silenceToFinal` | 700 ms | Consecutive low-energy audio required to end an utterance automatically. |
| `speechPad` | 300 ms | Maximum idle pre-roll normally retained before speech and maximum trailing silence kept in a silence-finalized chunk. |
| `overlap` | 500 ms | Tail retained across a forced split at `maxFinalDuration`. This is distinct from ordinary speech padding. |
| `partialWindow` | 5 s | Maximum amount of recent buffered audio copied into a partial chunk. Older samples remain available for the final chunk. |
| `partialInterval` | 2 s | Minimum buffered audio added since the previous partial before another partial is emitted. For the first partial, the measurement starts at the beginning of the utterance buffer, including pre-roll. |
| `maxFinalDuration` | 8 s | Maximum buffered utterance length before a forced final is emitted and overlap is retained for continuation. |
| `energyThreshold` | 0.01 RMS | Boundary between speech and silence classification. Higher values reject more quiet audio; lower values accept more background noise. |

Timing settings are evaluated on frame boundaries. For example, with 100 ms
frames a 250 ms minimum cannot be reached exactly: three speech frames provide
300 ms and satisfy it.

## Internal state terms

These fields explain the names used by the implementation and tests.

| Field | Meaning |
| --- | --- |
| `Config` | Timing and energy settings used for subsequent frames. |
| `sampleCursor` | Total number of non-empty input samples accepted so far. It is the absolute start offset of the next frame and is not reset between utterances. |
| `inSpeech` | Whether an utterance is currently open. It describes collection state, not necessarily the classification of the latest frame. It remains true during trailing silence until finalization. |
| `speechStart` | Absolute sample offset corresponding to index zero of `speechSamples`. It includes any prepended pre-roll or overlap. |
| `speechSamples` | Complete buffer for the open utterance: pre-roll or overlap, speech frames, and any retained silence. |
| `preRollSamples` | Idle audio saved for possible inclusion at the start of the next utterance. After a forced split it initially contains overlap; after silence finalization it contains the latest trailing padding. |
| `silenceSamples` | Number of samples in the current consecutive run of silence frames. A speech frame resets it to zero. |
| `activeSpeechSamples` | Cumulative number of samples from speech-classified frames in the current utterance. Silence, pre-roll, and carried overlap do not increment it. |
| `lastPartialAt` | Length of `speechSamples` when the previous partial was emitted. It is used to measure new buffered audio for `partialInterval`. |

## Helper and implementation terms

| Name | Meaning |
| --- | --- |
| `isSpeech` | Calculates a frame's RMS amplitude and compares it with `energyThreshold`. |
| `addSpeechFrame` | Opens an utterance if needed, appends a speech frame, and evaluates partial and forced-final boundaries. |
| `addSilenceFrame` | Updates idle pre-roll or appends trailing silence and evaluates the silence-final boundary. |
| `shouldEmitPartial` | Requires both `minSpeech` active speech and `partialInterval` new buffered audio. |
| `partialChunk` | Copies the latest `partialWindow` from the utterance buffer and calculates its absolute timestamps. |
| `finalChunk` | Optionally drops excess trailing samples, rejects utterances below `minSpeech`, copies the remaining buffer, and marks it final. |
| `dropTailSamples` | Number of samples removed from the end while building a silence-finalized chunk. It trims trailing silence beyond `speechPad`. |
| `resetAfterFinal` | Clears open-utterance counters and installs supplied padding or overlap as the next pre-roll. It does not reset `sampleCursor`. |
| `appendPreRoll` | Adds idle non-speech samples to the rolling pre-roll buffer. |
| `capPreRoll` | Keeps only the newest `speechPad` of ordinary idle pre-roll. |
| `samplesForDuration` | Converts a duration to a sample count using `durationSeconds * sampleRate`; conversion to `int` truncates fractional samples. |
| `samplesDuration` | Converts a sample count to elapsed time using `sampleCount / sampleRate`. |
| `tailSamples` | Returns an owned copy of up to the requested number of samples from the end of a buffer. |

## Default behavior examples

For continuous speech, partials are normally considered every 2 seconds. Each
partial contains no more than the most recent 5 seconds. At 8 seconds the
chunker emits a forced final, retains the last 500 ms as overlap, and starts a
new utterance if speech continues.

For speech followed by silence, the chunker waits for 700 ms of consecutive
silence. It then emits a final containing only the last 300 ms of that silence.
The excess 400 ms is excluded. If the stream ends first, `Flush` finalizes all
currently buffered samples without applying that trailing-silence trim.
