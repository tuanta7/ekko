import type { RecorderPhase } from "@/src/types/transcription.ts";

export function parsePhase(value: string): RecorderPhase {
  switch (value) {
    case "starting":
    case "recording":
    case "transcribing":
    case "stopping":
    case "stopped":
      return value;
    default:
      return "idle";
  }
}

export function labelState(phase: RecorderPhase): string {
  switch (phase) {
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

export function isActivePhase(phase: RecorderPhase): boolean {
  return phase === "starting" || phase === "recording" || phase === "transcribing" || phase === "stopping";
}
