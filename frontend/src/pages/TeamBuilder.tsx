import { useMemo, useState } from "react";
import { api, TeamPlan } from "../lib/api";
import { useApi } from "../lib/useApi";
import { StateGate } from "../components/StateGate";
import { PersonCard } from "../components/PersonCard";

export default function TeamBuilder() {
  const skills = useApi(() => api.skills(), []);
  const [selected, setSelected] = useState<string[]>([]);
  const [plan, setPlan] = useState<TeamPlan | null>(null);
  const [planning, setPlanning] = useState(false);
  const [planError, setPlanError] = useState<unknown>(null);

  const byCategory = useMemo(() => {
    const groups = new Map<string, { name: string; people: number }[]>();
    for (const s of skills.data ?? []) {
      const list = groups.get(s.category) ?? [];
      list.push({ name: s.name, people: s.people });
      groups.set(s.category, list);
    }
    return [...groups.entries()];
  }, [skills.data]);

  const toggle = (name: string) => {
    setPlan(null);
    setSelected((prev) =>
      prev.includes(name) ? prev.filter((s) => s !== name) : [...prev, name],
    );
  };

  const buildPlan = async () => {
    setPlanning(true);
    setPlanError(null);
    try {
      setPlan(await api.teamPlan(selected));
    } catch (err) {
      setPlanError(err);
    } finally {
      setPlanning(false);
    }
  };

  return (
    <>
      <h1>Team builder</h1>
      <p className="lede">
        Select the skills a new project needs. TalentGraph fetches every qualified
        person in one graph query, then assembles the smallest team that covers the
        whole list — strongest practitioners first.
      </p>

      <StateGate loading={skills.loading} error={skills.error} onRetry={skills.retry}>
        {byCategory.map(([category, list]) => (
          <div key={category} style={{ marginBottom: 14 }}>
            <div className="section-note" style={{ fontWeight: 700, textTransform: "uppercase", letterSpacing: "0.05em", fontSize: 11.5 }}>
              {category}
            </div>
            <div className="chips">
              {list.map((s) => (
                <button
                  key={s.name}
                  className={`chip ${selected.includes(s.name) ? "selected" : ""}`}
                  onClick={() => toggle(s.name)}
                  aria-pressed={selected.includes(s.name)}
                >
                  {s.name}
                </button>
              ))}
            </div>
          </div>
        ))}

        <div className="controls" style={{ marginTop: 24 }}>
          <button
            className="btn"
            disabled={selected.length === 0 || selected.length > 10 || planning}
            onClick={buildPlan}
          >
            {planning ? "Assembling…" : `Build team (${selected.length} skill${selected.length === 1 ? "" : "s"})`}
          </button>
          {selected.length > 0 && (
            <button className="btn ghost" onClick={() => { setSelected([]); setPlan(null); }}>
              Clear
            </button>
          )}
        </div>
      </StateGate>

      <StateGate loading={planning} error={planError} onRetry={buildPlan}>
        {plan && (
          <>
            <h2>
              Suggested team — {plan.team.length} {plan.team.length === 1 ? "person" : "people"}
            </h2>
            {plan.uncovered.length > 0 && (
              <div className="card" style={{ borderColor: "var(--critical)", marginBottom: 14 }}>
                <strong style={{ color: "var(--critical)" }}>Skill gap:</strong> no one
                in the graph covers {plan.uncovered.join(", ")}. Consider hiring or
                training.
              </div>
            )}
            <div className="grid cols-2">
              {plan.team.map((m) => (
                <PersonCard
                  key={m.person.id}
                  person={{ ...m.person, topSkills: [] }}
                  extra={
                    <div style={{ textAlign: "right", flexShrink: 0 }}>
                      {m.covers.map((c) => (
                        <div key={c.name} className="section-note">
                          <strong style={{ color: "var(--ink)" }}>{c.name}</strong>{" "}
                          L{c.level}
                        </div>
                      ))}
                    </div>
                  }
                />
              ))}
            </div>
            {plan.team.length === 0 && (
              <div className="state">
                <div className="title">No coverage at all</div>
                <p>None of the selected skills exist in the graph yet.</p>
              </div>
            )}
          </>
        )}
      </StateGate>
    </>
  );
}
