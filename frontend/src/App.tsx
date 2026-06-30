import { useEffect, useRef, useState } from "react";
import { Events } from "@wailsio/runtime";
import { TranscribeService } from "../bindings/github.com/tuanta7/ekko/services";
import type { ErrorEvent, StateEvent, TranscriptEvent } from "@/bindings/github.com/tuanta7/ekko/services";
import AppHeader from "./components/AppHeader";
import TranscriptMain from "./components/TranscriptMain";
import { useRecorder } from "./hooks/useRecorder";
import { isActivePhase } from "./lib/state";
import type { TranscriptLine } from "./types/transcription";

function App() {
  const [source, setSource] = useState("");
  const [sources, setSources] = useState<string[]>([]);
  const [partial, setPartial] = useState<TranscriptLine | null>(null);
  const [finalLines, setFinalLines] = useState<TranscriptLine[]>([]);

  const [recorder, dispatch] = useRecorder();

  const transcriptRef = useRef<HTMLDivElement>(null);

  const isStopping = recorder.phase === "stopping";
  const isActive = isActivePhase(recorder.phase);

  const refreshSources = () => {
    dispatch({ type: "refresh-requested" });
    TranscribeService.ListSources()
      .then((values: string[]) => {
        setSources(values);
        setSource((current) => (values.includes(current) ? current : values[0] || ""));
        dispatch({ type: "sources-loaded", count: values.length });
      })
      .catch((err: unknown) => {
        const message = String(err);
        setSources([]);
        setSource("");
        dispatch({ type: "sources-failed", message });
      });
  };

  useEffect(() => {
    refreshSources();
  }, []);

  useEffect(() => {
    const offState = Events.On("transcribe:state", (event: any) => {
      const data = event.data as StateEvent;
      dispatch({ type: "state-received", event: data });
      if (data.state === "stopped") {
        setPartial(null);
      }
    });

    const offPartial = Events.On("transcribe:partial", (event: any) => {
      const data = event.data as TranscriptEvent;
      setPartial(toLine(data));
    });

    const offFinal = Events.On("transcribe:final", (event: any) => {
      const data = event.data as TranscriptEvent;
      setFinalLines((current) => [...current, toLine(data)]);
      setPartial(null);
    });

    const offError = Events.On("transcribe:error", (event: any) => {
      dispatch({ type: "error-received", event: event.data as ErrorEvent });
    });

    return () => {
      offState();
      offPartial();
      offFinal();
      offError();
    };
  }, []);

  useEffect(() => {
    const frame = requestAnimationFrame(() => {
      const transcript = transcriptRef.current;
      if (transcript) {
        transcript.scrollTo({ top: transcript.scrollHeight, behavior: "smooth" });
      }
    });
    return () => cancelAnimationFrame(frame);
  }, [finalLines, partial, recorder.error]);

  const start = () => {
    setPartial(null);
    setFinalLines([]);
    dispatch({ type: "start-requested" });

    TranscribeService.Start(source)
      .then((sessionID: string) => {
        dispatch({ type: "start-resolved", sessionID });
      })
      .catch((err: unknown) => {
        dispatch({ type: "start-failed", message: String(err) });
      });
  };

  const stop = () => {
    if (!recorder.sessionID || isStopping) {
      return;
    }

    dispatch({ type: "stop-requested" });
    TranscribeService.Stop(recorder.sessionID).catch((err: unknown) => {
      dispatch({ type: "stop-failed", message: String(err) });
    });
  };

  const clearTranscript = () => {
    setFinalLines([]);
    setPartial(null);
  };

  const liveText = recorder.error || partial?.text || (isActive ? "Listening for speech…" : "No live segment yet.");

  return (
    <main className="h-full w-full bg-transparent font-sans text-gray-900 antialiased selection:bg-orange-200">
      <section
        className="mono-shell relative flex h-full w-full min-w-0 flex-col overflow-hidden"
        aria-label="Transcription controls"
      >
        <AppHeader
          recorder={recorder}
          source={source}
          sources={sources}
          hasTranscript={finalLines.length > 0 || Boolean(partial)}
          onSourceChange={setSource}
          onClear={clearTranscript}
          onRefresh={refreshSources}
          onStart={start}
          onStop={stop}
        />
        <TranscriptMain
          finalLines={finalLines}
          liveLine={partial}
          displayText={liveText}
          error={recorder.error}
          scrollContainerRef={transcriptRef}
        />
      </section>
    </main>
  );
}

function toLine(event: TranscriptEvent): TranscriptLine {
  return {
    id: event.chunkID,
    text: event.text,
    startMs: event.startMs,
    endMs: event.endMs,
  };
}

export default App;
