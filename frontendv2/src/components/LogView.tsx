import LogEntry, { LogEntryProps } from "./LogEntry";

export default function LogView({ logs }: { logs: LogEntryProps[] }) {
  return (
    <main className="flex-1 p-6 bg-panel-dark">
      <div className="flex flex-wrap justify-between gap-3 pb-4 border-b border-border-dark">
        <div className="flex min-w-72 flex-col gap-1">
          <p className="text-text-light text-2xl font-bold leading-tight tracking-[-0.033em]">Search Results</p>
          <p className="text-text-subtle-dark text-sm font-normal leading-normal">Showing {logs.length} results</p>
        </div>
      </div>
      <div className="flex flex-col mt-4 font-mono text-xs">
        {logs.map((log, index) => (
          <LogEntry key={index} {...log} />
        ))}
      </div>
    </main>
  );
}
