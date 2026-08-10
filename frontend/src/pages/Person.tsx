import { Link, useParams } from "react-router-dom";
import { api } from "../lib/api";
import { useApi } from "../lib/useApi";
import { StateGate } from "../components/StateGate";
import { LevelDots } from "../components/PersonCard";

export default function Person() {
  const { id = "" } = useParams();
  const person = useApi(() => api.person(id), [id]);
  const p = person.data;

  return (
    <StateGate loading={person.loading} error={person.error} onRetry={person.retry}>
      {p && (
        <>
          <h1>{p.name}</h1>
          <p className="lede">
            {p.title} · {p.seniority} · {p.location}
          </p>

          <div className="controls">
            <Link className="btn ghost" to={`/experts?from=${p.id}`}>
              Find experts near {p.name.split(" ")[0]}
            </Link>
            <Link className="btn ghost" to={`/path?from=${p.id}`}>
              Find an intro path from {p.name.split(" ")[0]}
            </Link>
          </div>

          <div className="grid cols-2" style={{ alignItems: "start" }}>
            <div className="card">
              <h2 style={{ marginTop: 0 }}>Skills</h2>
              {p.skills.length === 0 && (
                <p className="section-note">No skills recorded.</p>
              )}
              {p.skills.map((s) => (
                <div key={s.name} className="skill-row">
                  <span>
                    {s.name}
                    <span className="section-note" style={{ marginLeft: 8 }}>
                      {s.years} yr{s.years === 1 ? "" : "s"}
                    </span>
                  </span>
                  <LevelDots level={s.level} />
                </div>
              ))}
            </div>

            <div className="card">
              <h2 style={{ marginTop: 0 }}>Projects</h2>
              {p.projects.length === 0 && (
                <p className="section-note">No project history.</p>
              )}
              {p.projects.map((pr) => (
                <div key={pr.id} className="skill-row">
                  <span>
                    {pr.name}
                    <div className="section-note">
                      {pr.client} · {pr.role} · {pr.year}
                      {pr.status === "active" ? " · active" : ""}
                    </div>
                  </span>
                </div>
              ))}
            </div>

            <div className="card">
              <h2 style={{ marginTop: 0 }}>Frequent collaborators</h2>
              {p.colleagues.length === 0 && (
                <p className="section-note">No shared projects yet.</p>
              )}
              {p.colleagues.map((c) => (
                <div key={c.person.id} className="list-row">
                  <span>
                    <Link to={`/people/${c.person.id}`}>{c.person.name}</Link>
                    <div className="section-note">{c.person.title}</div>
                  </span>
                  <span className="badge near">
                    {c.sharedProjects} shared project{c.sharedProjects === 1 ? "" : "s"}
                  </span>
                </div>
              ))}
            </div>

            <div className="card">
              <h2 style={{ marginTop: 0 }}>Mentoring</h2>
              {p.mentor ? (
                <div className="list-row">
                  <span>
                    Mentored by{" "}
                    <Link to={`/people/${p.mentor.id}`}>{p.mentor.name}</Link>
                    <div className="section-note">{p.mentor.title}</div>
                  </span>
                </div>
              ) : (
                <p className="section-note">No mentor recorded.</p>
              )}
              {p.mentees.length > 0 && (
                <>
                  <p className="section-note" style={{ marginTop: 12 }}>
                    Mentors:
                  </p>
                  {p.mentees.map((m) => (
                    <div key={m.id} className="list-row">
                      <span>
                        <Link to={`/people/${m.id}`}>{m.name}</Link>
                        <div className="section-note">{m.title}</div>
                      </span>
                    </div>
                  ))}
                </>
              )}
            </div>
          </div>
        </>
      )}
    </StateGate>
  );
}
