import type { CSSProperties } from "react";
import { AlertCircle, CheckCircle2, GripVertical, Mic, Play, RefreshCw, Square, Trash2 } from "lucide-react";

import type { RecorderPhase, RecorderState } from "../types/transcription";
import {labelState} from "@/src/lib/state.ts";

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
    <header className="relative z-10 flex h-12 shrink-0 items-center gap-2 px-2">
      <div
        className="flex h-8 w-5 cursor-grab items-center justify-center rounded-lg text-zinc-500 transition hover:bg-zinc-300 hover:text-zinc-800 active:cursor-grabbing"
        style={{ "--wails-draggable": "drag" } as CSSProperties}
        title="Drag to move"
      >
        <GripVertical size={14} />
      </div>
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2 text-[10px] font-bold uppercase tracking-widest">
          <span className={`flex items-center gap-1.5 ${phaseColour(recorder.phase)}`}>
            <span className="relative flex h-1.5 w-1.5">
              {isActive && !isStopping && (
                <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-zinc-500 opacity-50" />
              )}
              <span className={`relative inline-flex h-1.5 w-1.5 rounded-full ${phaseDot(recorder.phase)}`} />
            </span>
            {labelState(recorder.phase)}
          </span>
        </div>
        <div className="mt-0.5 flex min-w-0 items-center gap-1 text-[11px] leading-tight">
          {recorder.error ? (
            <AlertCircle size={12} className="shrink-0 text-zinc-800" />
          ) : (
            <CheckCircle2 size={12} className="shrink-0 text-zinc-500" />
          )}
          <span className={`truncate ${recorder.error ? "text-zinc-900" : "text-zinc-600"}`}>{statusText}</span>
        </div>
      </div>

      <div className="flex shrink-0 items-center gap-2">
        <button
          type="button"
          onClick={onClear}
          disabled={!hasTranscript}
          className="mono-button grid h-8 w-8 place-items-center rounded-lg disabled:cursor-not-allowed disabled:opacity-30"
          title="Clear transcript"
          aria-label="Clear transcript"
        >
          <Trash2 size={13} />
        </button>

        <button
          type="button"
          onClick={onRefresh}
          disabled={isActive}
          className="mono-button grid h-8 w-8 place-items-center rounded-lg disabled:cursor-not-allowed disabled:opacity-40"
          title="Refresh audio sources"
          aria-label="Refresh audio sources"
        >
          <RefreshCw size={14} />
        </button>

        <div className="relative flex items-center">
          <Mic size={14} className="pointer-events-none absolute left-2.5 z-10 text-zinc-600" />
          <select
            id="audio-source"
            value={source}
            onChange={(event) => onSourceChange(event.target.value)}
            disabled={isActive}
            className="mono-select h-8 w-40 appearance-none rounded-lg pl-8 pr-3 font-medium outline-none disabled:cursor-not-allowed disabled:opacity-50 [&>option]:bg-zinc-100"
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
            className={`flex h-8 w-8 items-center justify-center rounded-lg disabled:cursor-not-allowed disabled:opacity-50 ${
              isStarting || isStopping ? "mono-button" : "mono-button-primary active:scale-95"
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
            className="mono-button-primary flex h-8 w-8 items-center justify-center rounded-lg active:scale-95 disabled:cursor-not-allowed disabled:opacity-35"
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

function isActivePhase(phase: RecorderPhase): boolean {
  return phase === "starting" || phase === "recording" || phase === "transcribing" || phase === "stopping";
}

function phaseColour(phase: RecorderPhase): string {
  return isActivePhase(phase) ? "text-zinc-800" : "text-zinc-500";
}

function phaseDot(phase: RecorderPhase): string {
  return isActivePhase(phase) ? "bg-zinc-700" : "bg-zinc-400";
}

export default AppHeader;
