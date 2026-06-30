import type { RefObject } from "react";

import { formatTimeFromMilliseconds as formatTime } from "../lib/format";
import type { TranscriptLine } from "../types/transcription";

type TranscriptMainProps = {
  finalLines: TranscriptLine[];
  liveLine: TranscriptLine | null;
  displayText: string;
  error: string;
  scrollContainerRef: RefObject<HTMLDivElement>;
};

function TranscriptMain({ finalLines, liveLine, displayText, error, scrollContainerRef }: TranscriptMainProps) {
  return (
    <div
      ref={scrollContainerRef}
      className="transcript-scroll relative z-10 min-h-0 flex-1 overflow-y-auto bg-white"
    >
      <div className="grid gap-2 p-3">
        {finalLines.map((line) => (
          <TranscriptSegment key={line.id} line={line} />
        ))}
        <article
          className={`grid gap-0.5 rounded px-3 py-2 ${
            error
              ? "bg-red-50"
              : liveLine
                ? "bg-orange-50 animate-pulse"
                : "bg-gray-50"
          }`}
          aria-live="polite"
        >
          {liveLine && (
            <time className="text-[9px] font-bold tabular-nums tracking-wide text-orange-700 uppercase">
              {formatTime(liveLine.startMs)} – {formatTime(liveLine.endMs)}
            </time>
          )}
          <p className={`text-sm leading-6 font-medium ${
            error
              ? "text-red-900"
              : liveLine
                ? "text-gray-900"
                : "text-gray-600"
          }`}>
            {displayText}
          </p>
        </article>
      </div>
    </div>
  );
}

function TranscriptSegment({ line }: { line: TranscriptLine }) {
  return (
    <article className="grid gap-0.5 rounded bg-gray-50 px-3 py-2">
      <time className="text-[9px] font-bold tabular-nums tracking-wide text-orange-600 uppercase">
        {formatTime(line.startMs)} – {formatTime(line.endMs)}
      </time>
      <p className="text-sm leading-6 text-gray-900">{line.text}</p>
    </article>
  );
}

export default TranscriptMain;
