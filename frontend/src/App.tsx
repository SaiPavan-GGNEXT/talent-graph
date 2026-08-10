import { NavLink, Route, Routes } from "react-router-dom";
import { useTheme } from "./lib/theme";
import Home from "./pages/Home";
import People from "./pages/People";
import Person from "./pages/Person";
import Experts from "./pages/Experts";
import PathFinder from "./pages/PathFinder";
import TeamBuilder from "./pages/TeamBuilder";

export default function App() {
  const { theme, toggle } = useTheme();
  return (
    <>
      <nav className="nav">
        <div className="nav-inner">
          <NavLink to="/" className="brand">
            <span className="brand-dot" aria-hidden="true" />
            TalentGraph
          </NavLink>
          <div className="nav-links">
            <NavLink to="/people">Directory</NavLink>
            <NavLink to="/experts">Find experts</NavLink>
            <NavLink to="/path">Intro paths</NavLink>
            <NavLink to="/team">Team builder</NavLink>
          </div>
          <button
            className="theme-toggle"
            onClick={toggle}
            aria-label={`Switch to ${theme === "dark" ? "light" : "dark"} theme`}
            title={`Switch to ${theme === "dark" ? "light" : "dark"} theme`}
          >
            {theme === "dark" ? "☀" : "☾"}
          </button>
        </div>
      </nav>
      <main className="shell">
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/people" element={<People />} />
          <Route path="/people/:id" element={<Person />} />
          <Route path="/experts" element={<Experts />} />
          <Route path="/path" element={<PathFinder />} />
          <Route path="/team" element={<TeamBuilder />} />
        </Routes>
        <footer className="footer">
          TalentGraph — an expertise &amp; collaboration graph on CognoDB. Built for the
          Wexa AI take-home assignment.
        </footer>
      </main>
    </>
  );
}
