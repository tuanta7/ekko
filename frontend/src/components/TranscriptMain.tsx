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
      className="transcript-scroll relative z-10 mx-2 mb-2 min-h-0 flex-1 overflow-y-auto rounded-2xl bg-zinc-200/60 px-2 py-2"
    >
      <div className="mx-auto grid gap-2">
        {finalLines.map((line) => (
          <TranscriptSegment key={line.id} line={line} />
        ))}
        <article className="grid gap-0.5 rounded-xl bg-zinc-100/80 px-2.5 py-1.5" aria-live="polite">
          {liveLine && (
            <time className="text-[10px] font-semibold tabular-nums tracking-wide text-zinc-600">
              {formatTime(liveLine.startMs)} – {formatTime(liveLine.endMs)}
            </time>
          )}
          <p className={`text-sm leading-5 ${error ? "text-zinc-950" : liveLine ? "text-zinc-900" : "text-zinc-500"}`}>
            {displayText}
          </p>
        </article>
      </div>
    </div>
  );
}

function TranscriptSegment({ line }: { line: TranscriptLine }) {
  return (
    <article className="grid gap-0.5 rounded-xl bg-zinc-100/80 px-2.5 py-1.5">
      <time className="text-[10px] font-semibold tabular-nums tracking-wide text-zinc-600">
        {formatTime(line.startMs)} – {formatTime(line.endMs)}
      </time>
      <p className="text-sm leading-5 text-zinc-800">{line.text}</p>
    </article>
  );
}

export default TranscriptMain;
