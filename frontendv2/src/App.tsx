import { useState, useEffect, useRef } from 'react';
import axios from 'axios';
import Header from './components/Header';
import Sidebar from './components/Sidebar';
import LogList from './components/LogList';
import { type LogEntry } from './types';

function App() {
  const [query, setQuery] = useState('');
  const [selectedLevels, setSelectedLevels] = useState<string[]>([]);
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [hasSearched, setHasSearched] = useState(false);
  const [isLiveTail, setIsLiveTail] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);

  // Handle Live Tail WebSocket
  useEffect(() => {
    if (isLiveTail) {
      // Clear logs when starting live tail
      setLogs([]);
      setError(null);
      setHasSearched(false);

      // Determine WebSocket URL (handle dev proxy vs prod)
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const host = window.location.host;
      // If running in dev with proxy, this works. If prod, this works.
      const wsUrl = `${protocol}//${host}/api/v1/tail`;

      console.log(`Connecting to WebSocket: ${wsUrl}`);
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        console.log('WebSocket connected');
      };

      ws.onmessage = (event) => {
        try {
          const logEntry: LogEntry = JSON.parse(event.data);
          setLogs(prevLogs => {
            // Keep only the last 1000 logs to prevent memory issues
            const newLogs = [logEntry, ...prevLogs];
            if (newLogs.length > 1000) {
              return newLogs.slice(0, 1000);
            }
            return newLogs;
          });
        } catch (e) {
          console.error('Failed to parse log entry:', e);
        }
      };

      ws.onerror = (e) => {
        console.error('WebSocket error:', e);
        setError('Live Tail connection error. Retrying...');
      };

      ws.onclose = () => {
        console.log('WebSocket disconnected');
        if (isLiveTail) {
          // If it closed unexpectedly, we might want to show an error or try to reconnect.
          // For now, let's just leave it.
        }
      };

      return () => {
        if (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING) {
          ws.close();
        }
        wsRef.current = null;
      };
    } else {
      // Cleanup if toggled off
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    }
  }, [isLiveTail]);

  const toggleLiveTail = () => {
    setIsLiveTail(!isLiveTail);
  };

  const handleSearch = async () => {
    if (isLiveTail) return; // Disable search if live tail is active

    if (!query.trim()) {
      setLogs([]);
      setError(null);
      return;
    }

    setHasSearched(true);
    setIsLoading(true);
    setError(null);
    try {
      // Construct the final query
      let finalQuery = query;
      if (selectedLevels.length > 0) {
        // If multiple levels, OR them: (level:ERROR OR level:WARN)
        // If single: level:ERROR
        const levelQuery = selectedLevels.map(l => `level:${l}`).join(' OR ');
        if (finalQuery) {
          finalQuery = `${finalQuery} AND (${levelQuery})`;
        } else {
          finalQuery = selectedLevels.length > 1 ? `(${levelQuery})` : levelQuery;
        }
      }

      // Use the same API endpoint as the original frontend
      const response = await axios.get<LogEntry[]>(`/api/v1/search?q=${encodeURIComponent(finalQuery)}&size=50`);
      setLogs(response.data || []);
    } catch (err) {
      console.error(err);
      setError('Failed to fetch logs. Is the backend running?');
      setLogs([]);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="relative flex h-screen w-full flex-col overflow-hidden bg-background-dark text-text-light font-display">
      <Header
        query={query}
        setQuery={setQuery}
        onSearch={handleSearch}
        isLiveTail={isLiveTail}
        onToggleLiveTail={toggleLiveTail}
      />
      <div className="flex flex-1 overflow-hidden">
        <Sidebar
          selectedLevels={selectedLevels}
          onLevelChange={(level, isChecked) => {
            setSelectedLevels(prev =>
              isChecked ? [...prev, level] : prev.filter(l => l !== level)
            );
          }}
        />
        <LogList
          logs={logs}
          isLoading={isLoading}
          error={error}
          hasSearched={hasSearched}
          isLiveTail={isLiveTail}
        />
      </div>
    </div>
  );
}

export default App;
