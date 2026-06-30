import type { CSSProperties } from "react";
import { AlertCircle, CheckCircle2, GripVertical, Mic, Play, RefreshCw, Square, Trash2 } from "lucide-react";

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
    <header className="relative z-10 flex h-14 shrink-0 items-center gap-3 bg-white px-4">
      <div
        className="flex h-8 w-6 cursor-grab items-center justify-center rounded-lg text-gray-600 transition hover:bg-gray-100 hover:text-gray-800 active:cursor-grabbing"
        style={{ "--wails-draggable": "drag" } as CSSProperties}
        title="Drag to move"
      >
        <GripVertical size={16} />
      </div>
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2 text-[11px] font-semibold uppercase tracking-widest">
          <span className={`flex items-center gap-2 ${phaseColour(recorder.phase)}`}>
            <span className="relative flex h-2 w-2">
              {isActive && !isStopping && (
                <span className="absolute inline-flex h-full w-full animate-ping rounded-full opacity-75" style={{ background: phaseColor(recorder.phase) }} />
              )}
              <span className={`relative inline-flex h-2 w-2 rounded-full ${phaseDot(recorder.phase)}`} />
            </span>
            {labelState(recorder.phase)}
          </span>
        </div>
        <div className="mt-1 flex min-w-0 items-center gap-1.5 text-[10px] leading-tight">
          {recorder.error ? (
            <AlertCircle size={13} className="shrink-0 text-red-600" />
          ) : (
            <CheckCircle2 size={13} className="shrink-0 text-green-600" />
          )}
          <span className={`truncate font-medium ${recorder.error ? "text-red-700" : "text-gray-600"}`}>{statusText}</span>
        </div>
      </div>

      <div className="flex shrink-0 items-center gap-2">
        <button
          type="button"
          onClick={onClear}
          disabled={!hasTranscript}
          className="cursor-pointer mono-button grid h-9 w-9 place-items-center rounded-lg disabled:cursor-not-allowed disabled:opacity-40"
          title="Clear transcript"
          aria-label="Clear transcript"
        >
          <Trash2 size={16} />
        </button>

        <button
          type="button"
          onClick={onRefresh}
          disabled={isActive}
          className="cursor-pointer mono-button grid h-9 w-9 place-items-center rounded-lg disabled:cursor-not-allowed disabled:opacity-40"
          title="Refresh audio sources"
          aria-label="Refresh audio sources"
        >
          <RefreshCw size={16} />
        </button>

        <div className="relative flex items-center">
          <Mic size={15} className="pointer-events-none absolute left-3 z-10 text-gray-600" />
          <select
            id="audio-source"
            value={source}
            onChange={(event) => onSourceChange(event.target.value)}
            disabled={isActive}
            className="cursor-pointer mono-select h-9 w-44 appearance-none rounded-lg pl-9 pr-3 outline-none disabled:cursor-not-allowed disabled:opacity-50"
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
            className={`flex cursor-pointer h-9 w-9 items-center justify-center rounded-lg disabled:cursor-not-allowed disabled:opacity-50 active:scale-95 ${
              isStarting || isStopping ? "mono-button" : "mono-button-primary"
            }`}
            title={isStarting ? "Starting recording" : isStopping ? "Stopping recording" : "Stop recording"}
            aria-label={isStarting ? "Starting recording" : isStopping ? "Stopping recording" : "Stop recording"}
          >
            {isStarting || isStopping ? (
              <RefreshCw size={14} className="animate-spin" />
            ) : (
              <Square size={14} fill="currentColor" />
            )}
          </button>
        ) : (
          <button
            type="button"
            onClick={onStart}
            disabled={!source}
            className="cursor-pointer mono-button-primary flex h-9 w-9 items-center justify-center rounded-lg active:scale-95 disabled:cursor-not-allowed disabled:opacity-40"
            title="Start recording"
            aria-label="Start recording"
          >
            <Play size={14} fill="currentColor" />
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
  return isActivePhase(phase) ? "text-orange-700" : "text-gray-500";
}

function phaseColor(phase: RecorderPhase): string {
  return isActivePhase(phase) ? "#FF8C42" : "#CCCCCC";
}

function phaseDot(phase: RecorderPhase): string {
  return isActivePhase(phase) ? "bg-orange-600" : "bg-gray-300";
}

export default AppHeader;
