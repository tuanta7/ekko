import { useEffect, useRef, useState } from "react";
import { Events } from "@wailsio/runtime";
import { StartOptions, TranscribeService } from "../bindings/github.com/tuanta7/ekko/services";
import type { ErrorEvent, TranscriptEvent } from "../bindings/github.com/tuanta7/ekko/services";
import type { StateEvent } from "../bindings/github.com/tuanta7/ekko/services/event";
import { Settings2, Play, Square, RefreshCw, Mic, GripVertical, AlertCircle, CheckCircle2 } from "lucide-react";
import { formatTimeFromMilliseconds as formatTime } from "./lib/format";

type FinalLine = {
  id: number;
  text: string;
  startMs: number;
  endMs: number;
};

function App() {
  const [source, setSource] = useState<string>("");
  const [sources, setSources] = useState<string[]>([]);

  const [modelName, setModelName] = useState<string>("ggml-base");
  const [language, setLanguage] = useState<string>("auto");
  const [sessionID, setSessionID] = useState<string>("");
  const [state, setState] = useState<string>("idle");
  const [status, setStatus] = useState<string>("Ready");
  const [error, setError] = useState<string>("");
  const [partial, setPartial] = useState<string>("");
  const [finalLines, setFinalLines] = useState<FinalLine[]>([]);
  const [settingsOpen, setSettingsOpen] = useState<boolean>(false);
  const transcriptRef = useRef<HTMLDivElement>(null);
  const stateRef = useRef<string>("idle");

  const isStarting = state === "starting";
  const isStopping = state === "stopping";
  const isRecording = state === "recording" || state === "transcribing";
  const isActive = isStarting || isRecording || isStopping;
  const canStop = isRecording && Boolean(sessionID);
  const liveText = error || partial || status;

  const refreshSources = () => {
    setError("");
    TranscribeService.ListSources()
      .then((values: string[]) => {
        setSources(values);
        setSource((current) => (values.includes(current) ? current : values[0] || ""));
        if (values.length === 0) {
          setStatus("No audio sources found");
        } else if (stateRef.current === "idle") {
          setStatus("Ready");
        }
      })
      .catch((err: any) => {
        setSources([]);
        setSource("");
        setStatus("Audio sources unavailable");
        setError(String(err));
      });
  };

  useEffect(() => {
    refreshSources();
  }, []);

  useEffect(() => {
    const offState = Events.On("transcribe:state", (event: any) => {
      const data = event.data as StateEvent;
      const nextState = data.state || "idle";
      if (stateRef.current === "stopping" && nextState !== "stopped") {
        return;
      }
      stateRef.current = nextState;
      setState(nextState);
      if (data.sessionID) {
        setSessionID(data.state === "stopped" ? "" : data.sessionID);
      }
      if (data.message) {
        setStatus(data.message);
      } else if (data.state) {
        setStatus(labelState(data.state));
      }
    });

    const offPartial = Events.On("transcribe:partial", (event: any) => {
      const data = event.data as TranscriptEvent;
      setPartial(data.text);
    });

    const offFinal = Events.On("transcribe:final", (event: any) => {
      const data = event.data as TranscriptEvent;
      setFinalLines((current) => [
        ...current,
        {
          id: data.chunkID,
          text: data.text,
          startMs: data.startMs,
          endMs: data.endMs,
        },
      ]);
      setPartial("");
    });

    const offError = Events.On("transcribe:error", (event: any) => {
      const data = event.data as ErrorEvent;
      setError(data.message);
      setStatus("Error");
    });

    return () => {
      offState();
      offPartial();
      offFinal();
      offError();
    };
  }, []);

  useEffect(() => {
    const transcript = transcriptRef.current;
    if (!transcript) {
      return;
    }
    transcript.scrollTop = transcript.scrollHeight;
  }, [finalLines, partial]);

  const start = () => {
    setError("");
    setPartial("");
    setFinalLines([]);
    stateRef.current = "starting";
    setState("starting");
    setStatus("Starting recorder");

    TranscribeService.Start(
      new StartOptions({
        source,
        modelName,
        language,
        threads: 0,
        translate: false,
      }),
    )
      .then((id: string) => {
        setSessionID(id);
      })
      .catch((err: any) => {
        stateRef.current = "idle";
        setState("idle");
        setStatus("Ready");
        setError(String(err));
      });
  };

  const stop = () => {
    if (!sessionID) {
      return;
    }

    stateRef.current = "stopping";
    setState("stopping");
    setStatus("Stopping recorder");
    TranscribeService.Stop(sessionID)
      .then(() => {
        stateRef.current = "stopped";
        setSessionID("");
        setState("stopped");
        setStatus("Recording stopped");
        setPartial("");
      })
      .catch((err: any) => {
        setError(String(err));
        stateRef.current = "recording";
        setState("recording");
        setStatus("Recording");
      });
  };

  return (
    <main className="h-full w-full bg-transparent text-white selection:bg-white/20 font-sans antialiased">
      <section
        className="relative flex h-full w-full min-w-0 flex-col overflow-hidden rounded-2xl border border-white/10 bg-zinc-950/80 shadow-2xl backdrop-blur-xl"
        aria-label="Transcription controls"
      >
        <header className="relative flex h-16 shrink-0 items-center gap-3 border-b border-white/10 px-3">
          <div
            className="flex h-10 w-6 cursor-grab items-center justify-center rounded-lg text-white/30 transition hover:bg-white/5 active:cursor-grabbing"
            style={{ "--wails-draggable": "drag" } as any}
            title="Drag to move"
          >
            <GripVertical size={18} />
          </div>

          <div className="relative shrink-0">
            <button
              type="button"
              onClick={() => setSettingsOpen((value) => !value)}
              className={`grid h-10 w-10 place-items-center rounded-lg transition-all duration-200 ${
                settingsOpen ? "bg-white/20 text-white" : "bg-white/5 text-white/70 hover:bg-white/10 hover:text-white"
              } focus:outline-none`}
              aria-expanded={settingsOpen}
              aria-label="Settings"
            >
              <Settings2 size={20} strokeWidth={1.5} />
            </button>

            {settingsOpen && (
              <div className="absolute left-0 top-full z-50 mt-3 w-80 overflow-hidden rounded-2xl border border-white/10 bg-zinc-950/95 p-4 shadow-2xl backdrop-blur-3xl">
                <div className="grid gap-4">
                  <div className="flex items-center justify-between">
                    <span className="text-xs font-semibold uppercase tracking-wider text-white/40">Configuration</span>
                    <button
                      onClick={refreshSources}
                      disabled={isActive}
                      className="flex items-center gap-1.5 rounded-md bg-white/5 px-2 py-1 text-[10px] font-medium text-white/60 transition hover:bg-white/10 hover:text-white disabled:opacity-30"
                    >
                      <RefreshCw size={12} className={isStarting ? "animate-spin" : ""} />
                      Refresh
                    </button>
                  </div>

                  <div className="grid gap-3">
                    <label className="grid gap-1.5">
                      <span className="text-[11px] font-medium text-white/40">Whisper Model</span>
                      <input
                        value={modelName}
                        onChange={(event) => setModelName(event.target.value)}
                        disabled={isActive}
                        placeholder="e.g. ggml-base"
                        className="h-9 w-full rounded-lg bg-white/5 px-3 text-sm text-white outline-none transition placeholder:text-white/20 focus:bg-white/10 disabled:opacity-50"
                      />
                    </label>

                    <label className="grid gap-1.5">
                      <span className="text-[11px] font-medium text-white/40">Target Language</span>
                      <input
                        value={language}
                        onChange={(event) => setLanguage(event.target.value)}
                        disabled={isActive}
                        placeholder="auto"
                        className="h-9 w-full rounded-lg bg-white/5 px-3 text-sm text-white outline-none transition placeholder:text-white/20 focus:bg-white/10 disabled:opacity-50"
                      />
                    </label>
                  </div>
                </div>
              </div>
            )}
          </div>

          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2 text-[10px] font-bold uppercase tracking-widest">
              {isActive ? (
                <span className={`flex items-center gap-1.5 ${isStopping ? "text-amber-300" : "text-emerald-400"}`}>
                  <span className="relative flex h-1.5 w-1.5">
                    <span
                      className={`absolute inline-flex h-full w-full animate-ping rounded-full ${isStopping ? "bg-amber-300" : "bg-emerald-400"} opacity-75`}
                    ></span>
                    <span
                      className={`relative inline-flex h-1.5 w-1.5 rounded-full ${isStopping ? "bg-amber-300" : "bg-emerald-400"}`}
                    ></span>
                  </span>
                  {labelState(state)}
                </span>
              ) : (
                <span className="flex items-center gap-1.5 text-white/30">
                  <span className="h-1.5 w-1.5 rounded-full bg-white/20"></span>
                  {labelState(state)}
                </span>
              )}
            </div>
            <div className="mt-1 flex min-w-0 items-center gap-1.5 truncate text-xs leading-tight">
              {error ? (
                <AlertCircle size={12} className="shrink-0 text-red-400" />
              ) : (
                <CheckCircle2 size={12} className="shrink-0 text-white/20" />
              )}
              <span className={`truncate ${error ? "text-red-400/90" : "text-white/50"}`}>{liveText}</span>
            </div>
          </div>

          <div className="flex shrink-0 items-center gap-3">
            <div className="relative flex items-center">
              <Mic size={14} className="absolute left-3 text-white/30" />
              <select
                id="audio-source"
                value={source}
                onChange={(event) => setSource(event.target.value)}
                disabled={isActive}
                className="h-10 w-48 appearance-none rounded-xl bg-white/5 pl-9 pr-4 text-xs font-medium text-white/80 outline-none transition hover:bg-white/10 focus:bg-white/10 disabled:cursor-not-allowed disabled:opacity-50 [&>option]:bg-zinc-900"
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
                onClick={stop}
                disabled={!canStop || isStopping}
                className={`flex h-10 w-10 items-center justify-center rounded-xl shadow-lg transition-all disabled:cursor-not-allowed disabled:opacity-60 ${
                  isStopping
                    ? "bg-amber-300 text-zinc-950 shadow-amber-300/10"
                    : "bg-white text-zinc-950 shadow-white/5 hover:scale-105 active:scale-95"
                }`}
                title={isStopping ? "Stopping Recording" : "Stop Recording"}
                aria-label={isStopping ? "Stopping recording" : "Stop recording"}
              >
                {isStopping ? (
                  <RefreshCw size={17} className="animate-spin" />
                ) : (
                  <Square size={18} fill="currentColor" />
                )}
              </button>
            ) : (
              <button
                type="button"
                onClick={start}
                disabled={!source}
                className="flex h-10 w-10 items-center justify-center rounded-xl bg-emerald-500 text-white shadow-lg shadow-emerald-500/20 transition-all hover:scale-105 hover:bg-emerald-400 active:scale-95 disabled:cursor-not-allowed disabled:opacity-30 disabled:hover:scale-100"
                title="Start Recording"
                aria-label="Start recording"
              >
                <Play size={18} fill="currentColor" className="ml-0.5" />
              </button>
            )}
          </div>
        </header>

        <div className="grid min-h-0 flex-1 grid-cols-[minmax(0,1fr)_18rem] gap-3 p-3">
          <section className="flex min-h-0 min-w-0 flex-col overflow-hidden rounded-lg border border-white/10 bg-black/20">
            <div className="flex h-10 shrink-0 items-center justify-between border-b border-white/10 px-3">
              <span className="text-[10px] font-bold uppercase tracking-widest text-white/40">Transcript</span>
              <span className="text-[10px] font-medium text-white/25">{finalLines.length} lines</span>
            </div>
            <div ref={transcriptRef} className="min-h-0 flex-1 overflow-y-auto px-3 py-2">
              {finalLines.length === 0 ? (
                <p className="py-8 text-center text-sm text-white/30">Waiting for transcript.</p>
              ) : (
                <div className="grid gap-2">
                  {finalLines.map((line) => (
                    <article key={line.id} className="grid gap-1 rounded-md bg-white/[0.03] px-3 py-2">
                      <time className="text-[10px] font-medium uppercase tracking-wider text-white/25">
                        {formatTime(line.startMs)} - {formatTime(line.endMs)}
                      </time>
                      <p className="text-sm leading-6 text-white/85">{line.text}</p>
                    </article>
                  ))}
                </div>
              )}
            </div>
          </section>

          <aside className="flex min-h-0 flex-col rounded-lg border border-white/10 bg-white/[0.04]">
            <div className="flex h-10 shrink-0 items-center border-b border-white/10 px-3">
              <span className="text-[10px] font-bold uppercase tracking-widest text-white/40">Live</span>
            </div>
            <div className="min-h-0 flex-1 overflow-y-auto p-3">
              <p
                className={`text-sm leading-6 ${error ? "text-red-300" : partial ? "text-white/90" : "text-white/35"}`}
              >
                {partial || error || "No live segment."}
              </p>
            </div>
          </aside>
        </div>
      </section>
    </main>
  );
}

function labelState(state: string): string {
  switch (state) {
    case "starting":
      return "Starting";
    case "recording":
      return "Recording";
    case "transcribing":
      return "Transcribing";
    case "stopping":
      return "Stopping";
    case "stopped":
      return "Stopped";
    default:
      return "Idle";
  }
}

export default App;
