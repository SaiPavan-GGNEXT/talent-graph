import { useState } from "react";
import { api } from "../lib/api";
import { useApi } from "../lib/useApi";
import { StateGate } from "../components/StateGate";
import { PersonCard } from "../components/PersonCard";

export default function People() {
  const [query, setQuery] = useState("");
  const people = useApi(() => api.people(query || undefined), [query]);

  return (
    <>
      <h1>Directory</h1>
      <p className="lede">Everyone in the graph, with their strongest skills.</p>

      <div className="controls">
        <div className="field">
          <label htmlFor="people-search">Search by name or title</label>
          <input
            id="people-search"
            type="search"
            placeholder="e.g. Priya, ML Engineer…"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
        </div>
      </div>

      <StateGate
        loading={people.loading}
        error={people.error}
        empty={people.data?.length === 0}
        emptyTitle="No people match"
        emptyHint="Try a different name or title fragment."
        onRetry={people.retry}
      >
        <div className="grid cols-2">
          {people.data?.map((p) => (
            <PersonCard key={p.id} person={p} />
          ))}
        </div>
      </StateGate>
    </>
  );
}
