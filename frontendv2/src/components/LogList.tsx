import React from 'react';
import { type LogEntry } from '../types';
import { format } from 'date-fns';

interface LogListProps {
    logs: LogEntry[];
    isLoading: boolean;
    error: string | null;
    hasSearched: boolean;
    isLiveTail: boolean;
}

const LogList: React.FC<LogListProps> = ({ logs, isLoading, error, hasSearched, isLiveTail }) => {
    const getLevelColor = (level: string) => {
        switch (level.toUpperCase()) {
            case 'ERROR': return 'bg-error text-white';
            case 'WARN': return 'bg-warning text-black';
            case 'INFO': return 'bg-info text-white';
            case 'DEBUG': return 'bg-debug text-white';
            default: return 'bg-panel-dark text-text-light';
        }
    };

    if (!hasSearched && !isLiveTail) {
        return (
            <main className="flex-1 p-6 bg-panel-dark overflow-y-auto flex items-center justify-center">
                <div className="text-center text-text-subtle-dark">
                    <p className="text-xl">Enter a query to search logs or start Live Tail</p>
                </div>
            </main>
        );
    }

    return (
        <main className="flex-1 p-6 bg-panel-dark overflow-y-auto">
            <div className="flex flex-wrap justify-between gap-3 pb-4 border-b border-border-dark">
                <div className="flex min-w-72 flex-col gap-1">
                    <div className="flex items-center gap-2">
                        <p className="text-text-light text-2xl font-bold leading-tight tracking-[-0.033em]">
                            {isLiveTail ? 'Live Logs' : 'Search Results'}
                        </p>
                        {isLiveTail && (
                            <span className="relative flex h-3 w-3">
                                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                                <span className="relative inline-flex rounded-full h-3 w-3 bg-green-500"></span>
                            </span>
                        )}
                    </div>
                    <p className="text-text-subtle-dark text-sm font-normal leading-normal">
                        {isLiveTail
                            ? `Streaming logs... (${logs.length} events)`
                            : (isLoading ? 'Searching...' : `Showing ${logs.length} results`)
                        }
                    </p>
                </div>
            </div>

            {error && (
                <div className="mt-4 p-4 bg-error/10 border border-error text-error rounded-md">
                    {error}
                </div>
            )}

            <div className="flex flex-col mt-4 font-mono text-xs">
                {logs.map((log, index) => (
                    <div key={index} className={`flex items-center gap-4 p-3 rounded-md ${index % 2 === 0 ? 'bg-row-odd-dark' : 'bg-panel-dark'}`}>
                        <span className="text-text-subtle-dark whitespace-nowrap">
                            {format(new Date(log.timestamp), 'yyyy-MM-dd HH:mm:ss.SSS')}
                        </span>
                        <span className={`px-2 py-1 text-xs font-bold rounded ${getLevelColor(log.level)}`}>
                            {log.level.toUpperCase()}
                        </span>
                        {log.labels && log.labels.service && (
                            <span className="px-2 py-1 text-xs font-medium rounded bg-primary/20 text-primary">
                                {log.labels.service}
                            </span>
                        )}
                        <p className="flex-1 text-text-light break-all">
                            {log.message}
                        </p>
                    </div>
                ))}

                {!isLoading && logs.length === 0 && !error && (
                    <div className="text-center text-text-subtle-dark py-10">
                        No logs found. Try adjusting your search or filters.
                    </div>
                )}
            </div>
        </main>
    );
};

export default LogList;
