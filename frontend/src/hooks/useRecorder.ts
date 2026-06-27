import { useReducer } from "react";

import type { ErrorEvent, StateEvent } from "@/bindings/github.com/tuanta7/ekko/services";

import { isActivePhase, labelState, parsePhase } from "../lib/state";
import type { RecorderState } from "../types/transcription";

type RecorderAction =
  | { type: "refresh-requested" }
  | { type: "sources-loaded"; count: number }
  | { type: "sources-failed"; message: string }
  | { type: "start-requested" }
  | { type: "start-resolved"; sessionID: string }
  | { type: "start-failed"; message: string }
  | { type: "stop-requested" }
  | { type: "stop-failed"; message: string }
  | { type: "state-received"; event: StateEvent }
  | { type: "error-received"; event: ErrorEvent };

const initialRecorderState: RecorderState = {
  sessionID: "",
  phase: "idle",
  status: "Ready",
  error: "",
};

export function useRecorder() {
  return useReducer(recorderReducer, initialRecorderState);
}

function recorderReducer(state: RecorderState, action: RecorderAction): RecorderState {
  switch (action.type) {
    case "refresh-requested":
      return { ...state, error: "" };
    case "sources-loaded":
      if (isActivePhase(state.phase)) {
        return state;
      }
      return { ...state, status: action.count > 0 ? "Ready" : "No audio sources found", error: "" };
    case "sources-failed":
      return { ...state, status: "Audio sources unavailable", error: action.message };
    case "start-requested":
      return { sessionID: "", phase: "starting", status: "Starting recorder", error: "" };
    case "start-resolved":
      if (state.phase === "stopped" || state.phase === "idle") {
        return state;
      }
      return {
        ...state,
        sessionID: action.sessionID,
        phase: state.phase === "starting" ? "recording" : state.phase,
        status: state.phase === "starting" ? "Recording started" : state.status,
      };
    case "start-failed":
      return { sessionID: "", phase: "idle", status: "Unable to start", error: action.message };
    case "stop-requested":
      return { ...state, phase: "stopping", status: "Stopping recorder", error: "" };
    case "stop-failed":
      if (state.phase !== "stopping") {
        return state;
      }
      return { sessionID: "", phase: "idle", status: "Unable to stop", error: action.message };
    case "state-received": {
      const nextPhase = parsePhase(action.event.state);
      if (state.sessionID && action.event.sessionID && state.sessionID !== action.event.sessionID && state.phase !== "starting") {
        return state;
      }
      if (state.phase === "stopping" && nextPhase !== "stopped") {
        return state;
      }
      return {
        sessionID: nextPhase === "stopped" ? "" : action.event.sessionID || state.sessionID,
        phase: nextPhase,
        status: action.event.message || labelState(nextPhase),
        error: nextPhase === "stopped" ? "" : state.error,
      };
    }
    case "error-received":
      if (state.sessionID && action.event.sessionID && state.sessionID !== action.event.sessionID) {
        return state;
      }
      return { ...state, status: "Transcription error", error: action.event.message };
  }
}
