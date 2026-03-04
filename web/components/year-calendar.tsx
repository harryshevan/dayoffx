import { buildMonthGrid } from "@/lib/calendar";
import { Vacation } from "@/lib/types";

const MONTH_NAMES = [
  "Январь",
  "Февраль",
  "Март",
  "Апрель",
  "Май",
  "Июнь",
  "Июль",
  "Август",
  "Сентябрь",
  "Октябрь",
  "Ноябрь",
  "Декабрь"
];

type YearCalendarProps = {
  year: number;
  vacations: Vacation[];
};

export function YearCalendar({ year, vacations }: YearCalendarProps) {
  return (
    <section className="year-grid">
      {MONTH_NAMES.map((monthName, monthIndex) => {
        const days = buildMonthGrid(year, monthIndex, vacations);
        return (
          <article key={monthName} className="month">
            <h3>{monthName}</h3>
            <div className="day-grid">
              {days.map((day) => (
                <div
                  key={day.date.toISOString()}
                  className={`day ${day.inCurrentMonth ? "" : "day-muted"}`}
                  title={day.vacationName ? `${day.vacationName}` : undefined}
                >
                  {day.day}
                  {day.vacationColor ? (
                    <span className="vac-dot" style={{ background: day.vacationColor }} />
                  ) : null}
                </div>
              ))}
            </div>
          </article>
        );
      })}
    </section>
  );
}
