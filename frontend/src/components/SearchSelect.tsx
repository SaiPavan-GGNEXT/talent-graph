import { useEffect, useMemo, useRef, useState } from "react";

export interface SelectOption {
  value: string;
  label: string;
  sub?: string;
}

// SearchSelect is a searchable combobox: type to filter, arrow keys to move,
// Enter to choose, Escape to close. Built plain so it can be walked through
// line by line — no component library.
export function SearchSelect({
  id,
  options,
  value,
  onChange,
  placeholder = "Type to search…",
}: {
  id: string;
  options: SelectOption[];
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
}) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [hilite, setHilite] = useState(0);
  const rootRef = useRef<HTMLDivElement>(null);
  const listRef = useRef<HTMLDivElement>(null);

  const chosen = options.find((o) => o.value === value) ?? null;

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return options;
    return options.filter(
      (o) =>
        o.label.toLowerCase().includes(q) || o.sub?.toLowerCase().includes(q),
    );
  }, [options, query]);

  // close on outside click
  useEffect(() => {
    const onDown = (e: MouseEvent) => {
      if (!rootRef.current?.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener("mousedown", onDown);
    return () => document.removeEventListener("mousedown", onDown);
  }, []);

  // keep the highlighted option scrolled into view
  useEffect(() => {
    listRef.current
      ?.querySelector(".hilite")
      ?.scrollIntoView({ block: "nearest" });
  }, [hilite]);

  const choose = (v: string) => {
    onChange(v);
    setQuery("");
    setOpen(false);
  };

  const onKeyDown = (e: React.KeyboardEvent) => {
    if (!open && (e.key === "ArrowDown" || e.key === "Enter")) {
      setOpen(true);
      return;
    }
    switch (e.key) {
      case "ArrowDown":
        e.preventDefault();
        setHilite((h) => Math.min(h + 1, filtered.length - 1));
        break;
      case "ArrowUp":
        e.preventDefault();
        setHilite((h) => Math.max(h - 1, 0));
        break;
      case "Enter":
        e.preventDefault();
        if (filtered[hilite]) choose(filtered[hilite].value);
        break;
      case "Escape":
        setOpen(false);
        break;
    }
  };

  return (
    <div className="combo" ref={rootRef}>
      <input
        id={id}
        type="text"
        role="combobox"
        aria-expanded={open}
        aria-autocomplete="list"
        placeholder={placeholder}
        value={open ? query : (chosen?.label ?? "")}
        onChange={(e) => {
          setQuery(e.target.value);
          setHilite(0);
          setOpen(true);
        }}
        onFocus={() => {
          setQuery("");
          setHilite(0);
          setOpen(true);
        }}
        onKeyDown={onKeyDown}
      />
      {chosen && !open && (
        <button
          type="button"
          className="combo-clear"
          aria-label="Clear selection"
          onClick={() => onChange("")}
        >
          ✕
        </button>
      )}
      {open && (
        <div className="combo-list" role="listbox" ref={listRef}>
          {filtered.length === 0 ? (
            <div className="combo-empty">No matches for “{query}”</div>
          ) : (
            filtered.map((o, i) => (
              <button
                type="button"
                key={o.value}
                role="option"
                aria-selected={o.value === value}
                className={`combo-option ${i === hilite ? "hilite" : ""} ${o.value === value ? "chosen" : ""}`}
                onMouseEnter={() => setHilite(i)}
                onClick={() => choose(o.value)}
              >
                {o.label}
                {o.sub && <span className="sub"> — {o.sub}</span>}
              </button>
            ))
          )}
        </div>
      )}
    </div>
  );
}
