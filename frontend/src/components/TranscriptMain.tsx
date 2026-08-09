import type { RefObject } from "react";

import { formatTimeFromMilliseconds as formatTime } from "../lib/format";
import type { TranscriptLine } from "../types/transcription";

type TranscriptMainProps = {
  finalLines: TranscriptLine[];
  liveLine: TranscriptLine | null;
  displayText: string;
  active: boolean;
  error: string;
  scrollContainerRef: RefObject<HTMLDivElement>;
};

function TranscriptMain({ finalLines, liveLine, displayText, active, error, scrollContainerRef }: TranscriptMainProps) {
  return (
    <div
      ref={scrollContainerRef}
      className="transcript-scroll relative z-10 min-h-0 flex-1 overflow-y-auto"
    >
      <div className="grid gap-1.5 p-2">
        {finalLines.map((line) => (
          <TranscriptSegment key={line.id} line={line} />
        ))}
        {(error || liveLine || active) && (
          <article
            className={`grid gap-0.5 rounded-md px-2.5 py-1.5 ${
              error ? "bg-red-500/15" : liveLine ? "bg-blue-500/15" : "bg-white/5"
            }`}
            aria-live="polite"
          >
            {liveLine && (
              <time className="text-[9px] font-bold tabular-nums tracking-wide text-blue-300 uppercase">
                {formatTime(liveLine.startMs)} – {formatTime(liveLine.endMs)}
              </time>
            )}
            {displayText ? (
              <p className={`text-[13px] leading-5 font-medium ${error ? "text-red-200" : "text-white"}`}>
                {displayText}
              </p>
            ) : (
              // Waiting on the first words of a chunk — dots stand in for the text.
              <span className="dots" aria-label="Waiting for speech">
                <i />
                <i />
                <i />
              </span>
            )}
          </article>
        )}
      </div>
    </div>
  );
}

function TranscriptSegment({ line }: { line: TranscriptLine }) {
  return (
    <article className="grid gap-0.5 rounded-md bg-white/5 px-2.5 py-1.5">
      <time className="text-[9px] font-bold tabular-nums tracking-wide text-blue-400 uppercase">
        {formatTime(line.startMs)} – {formatTime(line.endMs)}
      </time>
      <p className="text-[13px] leading-5 text-white/90">{line.text}</p>
    </article>
  );
}

export default TranscriptMain;
