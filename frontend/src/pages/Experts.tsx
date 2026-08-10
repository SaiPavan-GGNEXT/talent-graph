import { useSearchParams } from "react-router-dom";
import { api } from "../lib/api";
import { useApi } from "../lib/useApi";
import { StateGate } from "../components/StateGate";
import { PersonCard } from "../components/PersonCard";
import { SearchSelect } from "../components/SearchSelect";

function distanceBadge(distance: number, hasOrigin: boolean) {
  if (!hasOrigin) return null;
  if (distance === -1) return <span className="badge far">no shared path</span>;
  if (distance === 1) return <span className="badge direct">worked together</span>;
  return <span className="badge near">{distance} intros away</span>;
}

export default function Experts() {
  const [params, setParams] = useSearchParams();
  const skill = params.get("skill") ?? "";
  const from = params.get("from") ?? "";

  const skills = useApi(() => api.skills(), []);
  const people = useApi(() => api.people(), []);
  const experts = useApi(() => api.experts(skill, from || undefined), [skill, from], skill !== "");
  const adjacent = useApi(() => api.adjacentSkills(skill), [skill], skill !== "");

  const update = (key: string, value: string) => {
    const next = new URLSearchParams(params);
    if (value) next.set(key, value);
    else next.delete(key);
    setParams(next, { replace: true });
  };

  return (
    <>
      <h1>Find experts</h1>
      <p className="lede">
        Pick a skill to rank people by proficiency. Add a starting person and each
        expert is annotated with their collaboration distance — a multi-hop traversal
        through shared projects.
      </p>

      <div className="controls">
        <div className="field" style={{ minWidth: 280 }}>
          <label htmlFor="skill-select">Skill</label>
          <SearchSelect
            id="skill-select"
            value={skill}
            onChange={(v) => update("skill", v)}
            placeholder="Search skills…"
            options={(skills.data ?? []).map((s) => ({
              value: s.name,
              label: s.name,
              sub: `${s.category} · ${s.people} people`,
            }))}
          />
        </div>
        <div className="field" style={{ minWidth: 280 }}>
          <label htmlFor="from-select">Measure distance from (optional)</label>
          <SearchSelect
            id="from-select"
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
      </div>

      {!skill ? (
        <div className="state">
          <div className="title">Choose a skill to begin</div>
          <p>Experts are ranked by level and years of experience.</p>
        </div>
      ) : (
        <div className="grid" style={{ gridTemplateColumns: "2fr 1fr", alignItems: "start" }}>
          <div>
            <StateGate
              loading={experts.loading}
              error={experts.error}
              empty={experts.data?.length === 0}
              emptyTitle="No one holds this skill"
              emptyHint="Try an adjacent skill, or use the team builder to plan hiring."
              onRetry={experts.retry}
            >
              <div className="grid">
                {experts.data?.map((e) => (
                  <PersonCard
                    key={e.person.id}
                    person={e.person}
                    extra={
                      <div style={{ textAlign: "right", flexShrink: 0 }}>
                        <div style={{ fontWeight: 650 }}>
                          Level {e.level}/5
                          <span className="section-note" style={{ marginLeft: 6 }}>
                            {e.years} yrs
                          </span>
                        </div>
                        <div style={{ marginTop: 4 }}>
                          {distanceBadge(e.distance, from !== "")}
                        </div>
                      </div>
                    }
                  />
                ))}
              </div>
            </StateGate>
          </div>

          <div className="card">
            <h2 style={{ marginTop: 0 }}>Often paired with</h2>
            <p className="section-note">
              People who know {skill} usually also know…
            </p>
            <StateGate
              loading={adjacent.loading}
              error={adjacent.error}
              empty={adjacent.data?.length === 0}
              emptyTitle="No overlaps found"
            >
              {adjacent.data?.map((a) => (
                <div key={a.name} className="list-row">
                  <button
                    className="chip"
                    onClick={() => update("skill", a.name)}
                    title={`Switch to ${a.name}`}
                  >
                    {a.name}
                  </button>
                  <span className="section-note">{a.overlap} people</span>
                </div>
              ))}
            </StateGate>
          </div>
        </div>
      )}
    </>
  );
}
