export interface LogEntry {
    timestamp: string;
    level: string;
    message: string;
    labels: Record<string, string>;
}

export interface SearchResponse {
    data: LogEntry[];
    total: number;
}
