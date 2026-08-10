import { Link } from "react-router-dom";
import { PersonSummary } from "../lib/api";

export function PersonCard({
  person,
  extra,
}: {
  person: PersonSummary;
  extra?: React.ReactNode;
}) {
  return (
    <Link to={`/people/${person.id}`} className="card person-card">
      <div style={{ display: "flex", justifyContent: "space-between", gap: 10 }}>
        <div>
          <div className="person-name">{person.name}</div>
          <div className="person-meta">
            {person.title} · {person.location}
          </div>
        </div>
        {extra}
      </div>
      {person.topSkills?.length > 0 && (
        <div className="chips">
          {person.topSkills.map((s) => (
            <span key={s} className="chip">
              {s}
            </span>
          ))}
        </div>
      )}
    </Link>
  );
}

export function LevelDots({ level }: { level: number }) {
  return (
    <span className="level-dots" aria-label={`level ${level} of 5`}>
      {[1, 2, 3, 4, 5].map((i) => (
        <i key={i} className={i <= level ? "on" : ""} />
      ))}
    </span>
  );
}
