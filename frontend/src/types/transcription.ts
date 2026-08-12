export type RecorderPhase = "idle" | "starting" | "recording" | "transcribing" | "stopping" | "stopped";

export type RecorderState = {
  sessionID: string;
  phase: RecorderPhase;
  status: string;
  error: string;
};

export type TranscriptLine = {
  id: number;
  text: string;
  startMs: number;
  endMs: number;
};
