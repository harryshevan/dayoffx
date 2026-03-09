import { type CSSProperties, type MouseEvent, useEffect, useRef, useState } from "react";
import { buildMonthGrid, type DayCell, type DayVacation } from "@/lib/calendar";
import { VacationDots } from "@/components/vacation-dots";
import { Vacation } from "@/lib/types";

const MONTH_NAMES = [
  "January",
  "February",
  "March",
  "April",
  "May",
  "June",
  "July",
  "August",
  "September",
  "October",
  "November",
  "December"
];

const WEEKDAY_NAMES = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];

const POPOVER_WIDTH = 220;
const POPOVER_HEIGHT = 240;
const POPOVER_GAP = 12;
const POPOVER_PADDING = 8;

type YearCalendarProps = {
  year: number;
  vacations: Vacation[];
  highlightedMemberIds?: string[];
};

type DayPopover = {
  dateKey: string;
  left: number;
  top: number;
  vacations: DayVacation[];
};

function formatShortDate(value: string): string {
  const [, month, day] = value.split("-");
  if (!month || !day) {
    return value;
  }
  return `${day}.${month}`;
}

function formatVacationRange(item: DayVacation): string {
  return `${formatShortDate(item.fromDate)} -> ${formatShortDate(item.toDate)}`;
}

function getPopoverPosition(clientX: number, clientY: number): { left: number; top: number } {
  if (typeof window === "undefined") {
    return { left: clientX, top: clientY };
  }

  const maxLeft = Math.max(POPOVER_PADDING, window.innerWidth - POPOVER_WIDTH - POPOVER_PADDING);
  const left = Math.min(Math.max(POPOVER_PADDING, clientX + POPOVER_GAP), maxLeft);
  const wouldOverflowBottom = clientY + POPOVER_GAP + POPOVER_HEIGHT > window.innerHeight - POPOVER_PADDING;
  const top = wouldOverflowBottom
    ? Math.max(POPOVER_PADDING, clientY - POPOVER_HEIGHT - POPOVER_GAP)
    : clientY + POPOVER_GAP;

  return { left, top };
}

export function YearCalendar({ year, vacations, highlightedMemberIds = [] }: YearCalendarProps) {
  const today = new Date();
  const todayYear = today.getUTCFullYear();
  const todayMonth = today.getUTCMonth();
  const todayDay = today.getUTCDate();
  const [popover, setPopover] = useState<DayPopover | null>(null);
  const popoverRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    const handlePointerDown = (event: PointerEvent) => {
      if (!popover) {
        return;
      }
      if (popoverRef.current?.contains(event.target as Node)) {
        return;
      }
      setPopover(null);
    };

    document.addEventListener("pointerdown", handlePointerDown);
    return () => {
      document.removeEventListener("pointerdown", handlePointerDown);
    };
  }, [popover]);

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setPopover(null);
      }
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, []);

  useEffect(() => {
    setPopover(null);
  }, [year, vacations]);

  const handleDayClick = (event: MouseEvent<HTMLButtonElement>, day: DayCell) => {
    if (!day.vacations || day.vacations.length === 0) {
      return;
    }

    const dateKey = day.date.toISOString();
    if (popover?.dateKey === dateKey) {
      setPopover(null);
      return;
    }

    const { left, top } = getPopoverPosition(event.clientX, event.clientY);
    setPopover({
      dateKey,
      left,
      top,
      vacations: day.vacations
    });
  };

  return (
    <section className="year-grid">
      {MONTH_NAMES.map((monthName, monthIndex) => {
        const days = buildMonthGrid(year, monthIndex, vacations);
        return (
          <article key={monthName} className="month">
            <h3>{monthName}</h3>
            <div className="weekday-grid" aria-hidden="true">
              {WEEKDAY_NAMES.map((weekday) => (
                <div key={weekday} className="weekday-label">
                  {weekday}
                </div>
              ))}
            </div>
            <div className="day-grid">
              {days.map((day) => {
                const isToday =
                  day.inCurrentMonth &&
                  day.date.getUTCFullYear() === todayYear &&
                  day.date.getUTCMonth() === todayMonth &&
                  day.date.getUTCDate() === todayDay;
                const hasVacations = (day.vacations?.length ?? 0) > 0;
                const isPopoverOpen = popover?.dateKey === day.date.toISOString();
                const dayClassName = `day ${day.inCurrentMonth ? "" : "day-muted"} ${
                  highlightedMemberIds.length > 0 &&
                  day.vacationMemberIds?.some((memberId) => highlightedMemberIds.includes(memberId))
                    ? "day-highlighted"
                    : ""
                } ${isToday ? "day-today" : ""} ${hasVacations ? "day-clickable" : ""} ${
                  isPopoverOpen ? "day-popover-open" : ""
                }`;

                return (
                  hasVacations ? (
                    <button
                      key={day.date.toISOString()}
                      type="button"
                      className={dayClassName}
                      onClick={(event) => handleDayClick(event, day)}
                      aria-expanded={isPopoverOpen}
                      aria-label={`${day.day}: show vacations`}
                    >
                      {day.day}
                      <VacationDots dots={day.vacationDots} />
                    </button>
                  ) : (
                    <div key={day.date.toISOString()} className={dayClassName}>
                      {day.day}
                      <VacationDots dots={day.vacationDots} />
                    </div>
                  )
                );
              })}
            </div>
          </article>
        );
      })}
      {popover ? (
        <div
          ref={popoverRef}
          className="day-popover"
          style={{ left: `${popover.left}px`, top: `${popover.top}px` }}
          role="dialog"
          aria-label="Vacations for selected day"
        >
          <div className="day-popover-list">
            {popover.vacations.map((item, index) => (
              <div key={`${item.memberId}-${item.fromDate}-${item.toDate}-${index}`} className="day-popover-item">
                <span
                  className={`day-popover-dot ${item.status === "pending" ? "day-popover-dot-pending" : ""}`}
                  style={
                    item.status === "pending"
                      ? ({ "--vac-dot-color": item.colorHex } as CSSProperties)
                      : { background: item.colorHex }
                  }
                />
                <span className="day-popover-name">{item.displayName}</span>
                <span className="day-popover-range">{formatVacationRange(item)}</span>
              </div>
            ))}
          </div>
        </div>
      ) : null}
    </section>
  );
}
