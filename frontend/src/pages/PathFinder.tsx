import { useSearchParams } from "react-router-dom";
import { Link } from "react-router-dom";
import { api } from "../lib/api";
import { useApi } from "../lib/useApi";
import { StateGate } from "../components/StateGate";
import { SearchSelect } from "../components/SearchSelect";

const viaLabel: Record<string, string> = {
  WORKED_ON: "shared project",
  MENTORS: "mentoring",
};

export default function PathFinder() {
  const [params, setParams] = useSearchParams();
  const from = params.get("from") ?? "";
  const to = params.get("to") ?? "";

  const people = useApi(() => api.people(), []);
  const ready = from !== "" && to !== "" && from !== to;
  const path = useApi(() => api.path(from, to), [from, to], ready);

  const update = (key: string, value: string) => {
    const next = new URLSearchParams(params);
    if (value) next.set(key, value);
    else next.delete(key);
    setParams(next, { replace: true });
  };

  return (
    <>
      <h1>Intro paths</h1>
      <p className="lede">
        The shortest chain of shared projects and mentoring links between two people —
        so you can ask for a warm introduction instead of sending a cold message. This
        is a variable-length shortest-path query: one line of Cypher, and the kind of
        question relational databases struggle with.
      </p>

      <div className="controls">
        <div className="field" style={{ minWidth: 280 }}>
          <label htmlFor="from-person">From</label>
          <SearchSelect
            id="from-person"
            value={from}
            onChange={(v) => update("from", v)}
            placeholder="Search people…"
            options={(people.data ?? []).map((p) => ({
              value: p.id,
              label: p.name,
              sub: p.title,
            }))}
          />
        </div>
        <div className="field" style={{ minWidth: 280 }}>
          <label htmlFor="to-person">To</label>
          <SearchSelect
            id="to-person"
            value={to}
            onChange={(v) => update("to", v)}
            placeholder="Search people…"
            options={(people.data ?? []).map((p) => ({
              value: p.id,
              label: p.name,
              sub: p.title,
            }))}
          />
        </div>
      </div>

      {!ready ? (
        <div className="state">
          <div className="title">Pick two different people</div>
          <p>We’ll find the shortest chain of collaborations connecting them.</p>
        </div>
      ) : (
        <StateGate loading={path.loading} error={path.error} onRetry={path.retry}>
          {path.data &&
            (!path.data.connected || !path.data.path ? (
              <div className="state">
                <div className="title">No path found</div>
                <p>
                  These two people share no chain of projects or mentoring links (within
                  8 hops). A cold intro it is.
                </p>
              </div>
            ) : (
              <div className="card">
                <h2 style={{ marginTop: 0 }}>
                  Connected in {Math.floor(path.data.path.length / 2)} intro
                  {Math.floor(path.data.path.length / 2) === 1 ? "" : "s"}
                </h2>
                <div className="path-chain">
                  {path.data.path.map((node, i) => (
                    <div
                      key={`${node.type}-${node.id}`}
                      className={`path-node ${node.type}`}
                    >
                      <div className="path-marker">
                        <span className="dot" aria-hidden="true" />
                        {i < path.data!.path!.length - 1 && (
                          <span className="line" aria-hidden="true" />
                        )}
                      </div>
                      <div className="path-body">
                        {node.via && (
                          <div className="path-via">
                            via {viaLabel[node.via] ?? node.via.toLowerCase()}
                          </div>
                        )}
                        {node.type === "person" ? (
                          <>
                            <Link to={`/people/${node.id}`} style={{ fontWeight: 650 }}>
                              {node.name}
                            </Link>
                            <div className="section-note">{node.title}</div>
                          </>
                        ) : (
                          <>
                            <span style={{ fontWeight: 600 }}>{node.name}</span>
                            <div className="section-note">project</div>
                          </>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ))}
        </StateGate>
      )}
    </>
  );
}
