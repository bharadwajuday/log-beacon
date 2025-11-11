export interface LogEntryProps {
  timestamp: string;
  level: 'ERROR' | 'WARN' | 'INFO' | 'DEBUG';
  message: string;
}

const levelColors = {
  ERROR: 'bg-error',
  WARN: 'bg-warning',
  INFO: 'bg-info',
  DEBUG: 'bg-debug',
};

export default function LogEntry({ timestamp, level, message }: LogEntryProps) {
  return (
    <div className="flex items-center gap-4 p-3 bg-row-odd-dark rounded-md">
      <span className="text-text-subtle-dark">{timestamp}</span>
      <span className={`px-2 py-1 text-xs font-bold text-white rounded ${levelColors[level]}`}>{level}</span>
      <p className="flex-1 text-text-light">{message}</p>
    </div>
  );
}
