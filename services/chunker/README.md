# Audio Chunker

The chunker turns a continuous stream of mono float32 PCM samples into
overlapping `AudioChunk` values for transcription. It uses frame-level RMS
energy detection; it is not a word-level voice activity detector.

The service currently supplies 16 kHz audio in 100 ms frames. The chunker tracks
time from sample counts, so chunk timestamps remain tied to the original audio
stream.

## Process Flow

1. `AddFrame` ignores empty input, records the frame's position on the sample
   timeline, and classifies the complete frame as speech when its RMS energy is
   at least `0.01`.
2. While idle, silence is retained as a rolling 300 ms pre-roll buffer. When a
   speech frame arrives, the chunker starts an utterance and prepends that
   buffer so speech near the threshold boundary is not clipped.
3. While an utterance is active, every speech and silence frame is appended.
   Speech frames increase the active-speech duration; silence frames increase
   the trailing-silence duration.
4. Once the utterance contains at least 250 ms of active speech, the chunker may
   emit a partial every two seconds. A partial contains at most the latest five
   seconds, has `Final == false`, and does not reset the utterance.
5. After 700 ms of continuous silence, the chunker emits a final chunk when the
   utterance satisfies the 250 ms minimum. The final keeps up to 300 ms of
   trailing silence and drops any excess. The latest 300 ms is retained as
   pre-roll for the next utterance.
6. An utterance that reaches eight seconds is finalized even without trailing
   silence. The final contains the complete buffered utterance, and its latest
   500 ms becomes overlap for the next possible utterance.
7. `Flush` finalizes the current utterance without trimming its tail, provided
   it meets the minimum active-speech duration. It then resets the chunker
   without retaining pre-roll.

`AddFrame` can return both a partial and a final when their thresholds are
reached on the same frame; the partial is returned first. Chunks own copies of
their sample slices, and `Start`/`End` describe their positions in the original
sample timeline.

## Current Defaults

| Setting | Value | Effect |
| --- | ---: | --- |
| Sample rate | 16 kHz | Converts sample positions to durations |
| Energy threshold | 0.01 RMS | Classifies a frame as speech |
| Minimum speech | 250 ms | Rejects short noise bursts |
| Silence to final | 700 ms | Ends an utterance after continuous silence |
| Speech padding | 300 ms | Preserves boundary audio and trailing silence |
| Partial interval | 2 s | Controls preview frequency |
| Partial window | 5 s | Limits audio included in a partial chunk |
| Maximum final duration | 8 s | Bounds an utterance without silence |
| Maximum-duration overlap | 500 ms | Carries context into the next utterance |

These defaults are package-private fields in `DefaultConfig`; callers currently
use them through `NewAudioChunker` rather than configuring individual values.
