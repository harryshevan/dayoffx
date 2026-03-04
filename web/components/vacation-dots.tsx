type VacationDotsProps = {
  colors?: string[];
};

export function VacationDots({ colors }: VacationDotsProps) {
  const visibleColors = colors?.slice(0, 4) ?? [];
  const extraCount = Math.max((colors?.length ?? 0) - 4, 0);

  return (
    <>
      <span className="vac-dots">
        {visibleColors.map((color, index) => (
          <span key={`${color}-${index}`} className="vac-dot" style={{ background: color }} />
        ))}
      </span>
      {extraCount > 0 ? <span className="vac-more">+{extraCount}</span> : null}
    </>
  );
}
