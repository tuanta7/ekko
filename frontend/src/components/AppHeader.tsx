import type { CSSProperties } from "react";
import { AlertCircle, Circle, GripVertical, LoaderCircle, Mic, Play, RefreshCw, Square, Trash2 } from "lucide-react";

import type { RecorderPhase, RecorderState } from "../types/transcription";
import {labelState} from "../lib/state.ts";

type AppHeaderProps = {
  recorder: RecorderState;
  source: string;
  sources: string[];
  hasTranscript: boolean;
  onSourceChange: (source: string) => void;
  onClear: () => void;
  onRefresh: () => void;
  onStart: () => void;
  onStop: () => void;
};

function AppHeader({
  recorder,
  source,
  sources,
  hasTranscript,
  onSourceChange,
  onClear,
  onRefresh,
  onStart,
  onStop,
}: AppHeaderProps) {
  const isStarting = recorder.phase === "starting";
  const isStopping = recorder.phase === "stopping";
  const isActive = isActivePhase(recorder.phase);
  const canStop = Boolean(recorder.sessionID) &&
    (recorder.phase === "recording" || recorder.phase === "transcribing");
  const statusText = recorder.error || recorder.status;

  return (
    <header className="relative z-10 flex h-11 shrink-0 items-center gap-2 px-2.5">
      <div
        className="flex h-7 w-5 cursor-grab items-center justify-center rounded-lg text-white/40 transition hover:bg-white/10 hover:text-white/80 active:cursor-grabbing"
        style={{ "--wails-draggable": "drag" } as CSSProperties}
        title="Drag to move"
      >
        <GripVertical size={14} />
      </div>
      <div
        className={`flex min-w-0 flex-1 items-center ${phaseColour(recorder.phase)}`}
        role="status"
        title={statusText || labelState(recorder.phase)}
      >
        <PhaseSymbol phase={recorder.phase} error={Boolean(recorder.error)} />
        <span className="sr-only">{statusText || labelState(recorder.phase)}</span>
      </div>

      <div className="flex shrink-0 items-center gap-2">
        <button
          type="button"
          onClick={onClear}
          disabled={!hasTranscript}
          className="cursor-pointer mono-button grid h-7 w-7 place-items-center rounded-md disabled:cursor-not-allowed disabled:opacity-40"
          title="Clear transcript"
          aria-label="Clear transcript"
        >
          <Trash2 size={14} />
        </button>

        <button
          type="button"
          onClick={onRefresh}
          disabled={isActive}
          className="cursor-pointer mono-button grid h-7 w-7 place-items-center rounded-md disabled:cursor-not-allowed disabled:opacity-40"
          title="Refresh audio sources"
          aria-label="Refresh audio sources"
        >
          <RefreshCw size={14} />
        </button>

        <div className="relative flex items-center">
          <Mic size={13} className="pointer-events-none absolute left-2 z-10 text-white/50" />
          <select
            id="audio-source"
            value={source}
            onChange={(event) => onSourceChange(event.target.value)}
            disabled={isActive}
            className="cursor-pointer mono-select h-7 w-32 appearance-none rounded-md pl-7 pr-2 outline-none disabled:cursor-not-allowed disabled:opacity-50"
          >
            {sources.length === 0 && <option value="">No source found</option>}
            {sources.map((value) => (
              <option key={value} value={value}>
                {value}
              </option>
            ))}
          </select>
        </div>

        {isActive ? (
          <button
            type="button"
            onClick={onStop}
            disabled={!canStop}
            className={`flex cursor-pointer h-7 w-7 items-center justify-center rounded-md disabled:cursor-not-allowed disabled:opacity-50 active:scale-95 ${
              isStarting || isStopping ? "mono-button" : "mono-button-primary"
            }`}
            title={isStarting ? "Starting recording" : isStopping ? "Stopping recording" : "Stop recording"}
            aria-label={isStarting ? "Starting recording" : isStopping ? "Stopping recording" : "Stop recording"}
          >
            {isStarting || isStopping ? (
              <RefreshCw size={12} className="animate-spin" />
            ) : (
              <Square size={12} fill="currentColor" />
            )}
          </button>
        ) : (
          <button
            type="button"
            onClick={onStart}
            disabled={!source}
            className="cursor-pointer mono-button-primary flex h-7 w-7 items-center justify-center rounded-md active:scale-95 disabled:cursor-not-allowed disabled:opacity-40"
            title="Start recording"
            aria-label="Start recording"
          >
            <Play size={12} fill="currentColor" />
          </button>
        )}
      </div>
    </header>
  );
}

function PhaseSymbol({ phase, error }: { phase: RecorderPhase; error: boolean }) {
  if (error) {
    return <AlertCircle size={13} className="shrink-0 text-red-400" />;
  }

  switch (phase) {
    case "recording":
      // Bars wave in place — the only phase where audio is actually flowing in.
      return (
        <span className="wave" aria-hidden="true">
          <i />
          <i />
          <i />
          <i />
        </span>
      );
    case "transcribing":
      return <LoaderCircle size={13} className="shrink-0 animate-spin" />;
    case "starting":
    case "stopping":
      return <LoaderCircle size={13} className="shrink-0 animate-spin opacity-60" />;
    case "stopped":
      return null;
    default:
      return <Circle size={13} className="shrink-0 opacity-50" />;
  }
}

function isActivePhase(phase: RecorderPhase): boolean {
  return phase === "starting" || phase === "recording" || phase === "transcribing" || phase === "stopping";
}

function phaseColour(phase: RecorderPhase): string {
  return isActivePhase(phase) ? "text-blue-300" : "text-white/50";
}

export default AppHeader;
