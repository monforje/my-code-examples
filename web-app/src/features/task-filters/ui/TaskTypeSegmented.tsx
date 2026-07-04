import { Icon } from "@iconify/react";
import { SegmentedControl } from "@shared/ui";

interface TaskTypeSegmentedProps {
  value: string;
  onChange: (value: string) => void;
}

export function TaskTypeSegmented({ value, onChange }: TaskTypeSegmentedProps) {
  const options = [
    { value: "all", label: <TypeLabel icon="lucide:layout-grid" text="Все" /> },
    { value: "backend", label: <TypeLabel icon="lucide:server" text="Backend" /> },
    { value: "frontend", label: <TypeLabel icon="lucide:monitor" text="Frontend" /> },
  ];

  return <SegmentedControl value={value} options={options} onChange={onChange} />;
}

function TypeLabel({ icon, text }: { icon: string; text: string }) {
  return (
    <span className="segmented-type-label">
      <Icon icon={icon} width={16} height={16} />
      <span>{text}</span>
    </span>
  );
}
