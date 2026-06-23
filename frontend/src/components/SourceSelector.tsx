import { useEffect, useState, type Dispatch, type SetStateAction } from "react";

import { TranscribeService } from "../../bindings/github.com/tuanta7/ekko/services";

type SourceSelectorProps = {
  source: string;
  setSource: Dispatch<SetStateAction<string>>;
  disabled?: boolean;
};

const SourceSelector = ({ source, setSource, disabled = false }: SourceSelectorProps) => {
  const [sources, setSources] = useState<string[]>([]);
  const [error, setError] = useState<string>("");

  useEffect(() => {
    refreshSources();
  }, []);

  function refreshSources() {
    setError("");
    TranscribeService.ListSources()
      .then((values: string[]) => {
        setSources(values);
        if (!values.includes(source)) {
          setSource(values[0] || "");
        }
      })
      .catch((err: any) => {
        setError(String(err));
      });
  }

  return (
    <div className="grid gap-2">
      <div className="flex items-center gap-2">
        <select
          value={source}
          onChange={(event) => setSource(event.target.value)}
          disabled={disabled}
          className="min-w-0 flex-1 rounded-lg bg-white/5 px-3 py-2 text-sm text-white outline-none disabled:opacity-50"
        >
          {sources.length === 0 && <option value="">No source found</option>}
          {sources.map((value) => (
            <option key={value} value={value}>
              {value}
            </option>
          ))}
        </select>
        <button
          type="button"
          onClick={refreshSources}
          disabled={disabled}
          className="rounded-lg bg-white/10 px-3 py-2 text-xs font-medium text-white/70 disabled:opacity-50"
        >
          Refresh
        </button>
      </div>
      {error && <p className="text-xs text-red-300">{error}</p>}
    </div>
  );
};

export default SourceSelector;
