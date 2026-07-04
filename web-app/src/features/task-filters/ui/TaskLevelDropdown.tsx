import { Dropdown } from "@shared/ui";

const LEVEL_OPTIONS = [
  { value: "all", label: "Любой", dotColor: "var(--cd-text-subtle)" },
  { value: "junior", label: "База", dotColor: "#22c55e" },
  { value: "middle", label: "Практика", dotColor: "#f59e0b" },
  { value: "senior", label: "Сложно", dotColor: "#ef4444" },
];

interface TaskLevelDropdownProps {
  value: string;
  onChange: (value: string) => void;
}

export function TaskLevelDropdown({ value, onChange }: TaskLevelDropdownProps) {
  return (
    <Dropdown value={value} options={LEVEL_OPTIONS} placeholder="Уровень" onChange={onChange} />
  );
}
