import React from 'react';

interface SidebarProps {
    selectedLevels: string[];
    onLevelChange: (level: string, isChecked: boolean) => void;
}

const Sidebar: React.FC<SidebarProps> = ({ selectedLevels, onLevelChange }) => {
    const handleCheckboxChange = (level: string, e: React.ChangeEvent<HTMLInputElement>) => {
        onLevelChange(level, e.target.checked);
    };

    return (
        <aside className="flex h-full flex-col justify-between border-r border-border-dark bg-background-dark p-4 w-64">
            <div className="flex flex-col gap-4">
                <h3 className="text-lg font-bold px-2">Filters</h3>
                <div className="flex flex-col">
                    <details className="flex flex-col border-t border-t-border-dark py-2 group">
                        <summary className="flex cursor-pointer items-center justify-between gap-6 py-2 list-none">
                            <p className="text-text-light text-sm font-medium leading-normal">Time Range</p>
                            <span className="material-symbols-outlined text-text-light group-open:rotate-180 transition-transform">expand_more</span>
                        </summary>
                        <div className="flex flex-col gap-2 pt-2 px-2">
                            <button className="w-full text-left rounded-md p-2 text-sm bg-primary/20 text-primary">Last 15 minutes</button>
                            <button className="w-full text-left rounded-md p-2 text-sm hover:bg-panel-dark">Last hour</button>
                            <button className="w-full text-left rounded-md p-2 text-sm hover:bg-panel-dark">Last 24 hours</button>
                            <button className="w-full text-left rounded-md p-2 text-sm hover:bg-panel-dark">Custom...</button>
                        </div>
                    </details>
                    <details className="flex flex-col border-t border-t-border-dark py-2 group" open>
                        <summary className="flex cursor-pointer items-center justify-between gap-6 py-2 list-none">
                            <p className="text-text-light text-sm font-medium leading-normal">Log Level</p>
                            <span className="material-symbols-outlined text-text-light group-open:rotate-180 transition-transform">expand_more</span>
                        </summary>
                        <div className="px-2">
                            {['ERROR', 'WARN', 'INFO', 'DEBUG'].map((level) => (
                                <label key={level} className="flex gap-x-3 py-2 flex-row items-center">
                                    <input
                                        checked={selectedLevels.includes(level)}
                                        onChange={(e) => handleCheckboxChange(level, e)}
                                        className="h-5 w-5 rounded border-border-dark border-2 bg-transparent text-primary checked:bg-primary checked:border-primary checked:bg-[image:var(--checkbox-tick-svg)] focus:ring-0 focus:ring-offset-0 focus:border-border-dark focus:outline-none appearance-none"
                                        type="checkbox"
                                    />
                                    <p className="text-text-light text-sm font-normal leading-normal">{level}</p>
                                </label>
                            ))}
                        </div>
                    </details>

                </div>
            </div>
        </aside>
    );
};

export default Sidebar;
