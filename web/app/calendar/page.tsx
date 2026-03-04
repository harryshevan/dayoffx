"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { getVacations } from "@/lib/api";
import { Vacation } from "@/lib/types";
import { YearCalendar } from "@/components/year-calendar";

function getCurrentYear(): number {
  return new Date().getUTCFullYear();
}

export default function CalendarPage() {
  const currentYear = getCurrentYear();
  const [year, setYear] = useState(currentYear);
  const [vacations, setVacations] = useState<Vacation[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [hoveredMemberId, setHoveredMemberId] = useState<string | null>(null);
  const [selectedMemberId, setSelectedMemberId] = useState<string | null>(null);
  const legendRef = useRef<HTMLElement | null>(null);

  const yearOptions = useMemo(
    () => Array.from({ length: 5 }, (_, index) => currentYear - 2 + index),
    [currentYear]
  );

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);

    getVacations(year)
      .then((data) => {
        if (!cancelled) {
          setVacations(data);
        }
      })
      .catch((requestError) => {
        if (!cancelled) {
          setError(requestError instanceof Error ? requestError.message : "Failed to load vacations");
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [year]);

  const legendByMember = Array.from(
    vacations.reduce<Map<string, { memberId: string; name: string; color: string }>>((acc, item) => {
      if (!acc.has(item.memberId)) {
        acc.set(item.memberId, { memberId: item.memberId, name: item.displayName, color: item.colorHex });
      }
      return acc;
    }, new Map()).values()
  );
  const highlightedMemberId = selectedMemberId ?? hoveredMemberId;

  useEffect(() => {
    const handlePointerDown = (event: PointerEvent) => {
      if (!selectedMemberId) {
        return;
      }
      if (legendRef.current?.contains(event.target as Node)) {
        return;
      }
      setSelectedMemberId(null);
      setHoveredMemberId(null);
    };

    document.addEventListener("pointerdown", handlePointerDown);
    return () => {
      document.removeEventListener("pointerdown", handlePointerDown);
    };
  }, [selectedMemberId]);

  return (
    <main className="grid">
      <section className="card">
        <div style={{ display: "flex", justifyContent: "space-between", gap: "0.75rem", flexWrap: "wrap" }}>
          <h1 style={{ margin: 0 }}>Годовой календарь отпусков</h1>
          <label>
            Год{" "}
            <select value={year} onChange={(event) => setYear(Number(event.target.value))}>
              {yearOptions.map((optionYear) => (
                <option key={optionYear} value={optionYear}>
                  {optionYear}
                </option>
              ))}
            </select>
          </label>
        </div>
      </section>

      {legendByMember.length > 0 ? (
        <section ref={legendRef} className="legend">
          {legendByMember.map((item) => (
            <button
              key={item.memberId}
              type="button"
              className={`legend-item ${highlightedMemberId === item.memberId ? "legend-item-active" : ""}`}
              onMouseEnter={() => setHoveredMemberId(item.memberId)}
              onMouseLeave={() => setHoveredMemberId(null)}
              onClick={() => setSelectedMemberId((previous) => (previous === item.memberId ? null : item.memberId))}
              aria-pressed={selectedMemberId === item.memberId}
            >
              <span className="dot" style={{ background: item.color }} />
              {item.name}
            </button>
          ))}
        </section>
      ) : null}

      {error ? <section className="card">Ошибка: {error}</section> : null}
      {loading ? <section className="card">Загружаем календарь...</section> : null}
      {!loading && !error ? (
        <YearCalendar year={year} vacations={vacations} highlightedMemberId={highlightedMemberId} />
      ) : null}
    </main>
  );
}
