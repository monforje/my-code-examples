import { SearchInput } from "@shared/ui";

interface TaskSearchInputProps {
  value: string;
  onChange: (value: string) => void;
}

export function TaskSearchInput({ value, onChange }: TaskSearchInputProps) {
  return <SearchInput value={value} onChange={onChange} />;
}
