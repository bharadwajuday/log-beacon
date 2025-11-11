export default function Sidebar() {
  return (
    <aside className="flex h-full flex-col justify-between border-r border-border-dark bg-background-dark p-4 w-64">
      <div className="flex flex-col gap-4">
        <h3 className="text-lg font-bold px-2">Filters</h3>
        <div className="flex flex-col">
          <details className="flex flex-col border-t border-t-border-dark py-2 group">
            <summary className="flex cursor-pointer items-center justify-between gap-6 py-2">
              <p className="text-text-light text-sm font-medium leading-normal">Time Range</p>
              <span className="material-symbols-outlined text-text-light group-open:rotate-180 transition-transform">expand_more</span>
            </summary>
          </details>
          <details className="flex flex-col border-t border-t-border-dark py-2 group">
            <summary className="flex cursor-pointer items-center justify-between gap-6 py-2">
              <p className="text-text-light text-sm font-medium leading-normal">Log Level</p>
              <span className="material-symbols-outlined text-text-light group-open:rotate-180 transition-transform">expand_more</span>
            </summary>
          </details>
          <details className="flex flex-col border-t border-b border-border-dark py-2 group">
            <summary className="flex cursor-pointer items-center justify-between gap-6 py-2">
              <p className="text-text-light text-sm font-medium leading-normal">Source</p>
              <span className="material-symbols-outlined text-text-light group-open:rotate-180 transition-transform">expand_more</span>
            </summary>
          </details>
        </div>
      </div>
    </aside>
  );
}
