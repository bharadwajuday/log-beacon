import { useState, useEffect, useRef } from 'react';
import axios from 'axios';
import Header from './components/Header';
import Sidebar from './components/Sidebar';
import LogList from './components/LogList';
import Auth from './components/Auth';
import { type LogEntry } from './types';

function App() {
  const [token, setToken] = useState<string | null>(localStorage.getItem('token'));
  const [hasUsers, setHasUsers] = useState<boolean>(true); // Assume true until checked
  const [isAuthLoading, setIsAuthLoading] = useState(true);

  const [query, setQuery] = useState('');
  const [selectedLevels, setSelectedLevels] = useState<string[]>([]);
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [hasSearched, setHasSearched] = useState(false);
  const [isLiveTail, setIsLiveTail] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);

  // Check system auth status and token validity
  useEffect(() => {
    const checkAuthStatus = async () => {
      try {
        const response = await axios.get('/api/v1/auth/status');
        setHasUsers(response.data.has_users);
      } catch (err) {
        console.error('Failed to check auth status:', err);
      } finally {
        setIsAuthLoading(false);
      }
    };
    checkAuthStatus();
  }, []);

  const handleLogin = (newToken: string) => {
    localStorage.setItem('token', newToken);
    setToken(newToken);
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    setToken(null);
    setLogs([]);
    setIsLiveTail(false);
  };

  // Handle Live Tail WebSocket
  useEffect(() => {
    if (isLiveTail && token) {
      // Clear logs when starting live tail
      setLogs([]);
      setError(null);
      setHasSearched(false);

      // Determine WebSocket URL
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const host = window.location.host;
      // Pass token as query param for WebSocket auth
      const wsUrl = `${protocol}//${host}/api/v1/tail?token=${token}`;

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

      ws.onclose = (event) => {
        console.log('WebSocket disconnected', event.code);
        if (event.code === 1008 || event.code === 3000) { // Policy Violation or custom Unauth
             handleLogout();
        }
      };

      return () => {
        if (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING) {
          ws.close();
        }
        wsRef.current = null;
      };
    } else {
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    }
  }, [isLiveTail, token]);

  const toggleLiveTail = () => {
    setIsLiveTail(!isLiveTail);
  };

  const handleSearch = async () => {
    if (isLiveTail) return;

    if (!query.trim()) {
      setLogs([]);
      setError(null);
      return;
    }

    setHasSearched(true);
    setIsLoading(true);
    setError(null);
    try {
      let finalQuery = query;
      if (selectedLevels.length > 0) {
        const levelQuery = selectedLevels.map(l => `level:${l}`).join(' OR ');
        if (finalQuery) {
          finalQuery = `${finalQuery} AND (${levelQuery})`;
        } else {
          finalQuery = selectedLevels.length > 1 ? `(${levelQuery})` : levelQuery;
        }
      }

      const response = await axios.get<LogEntry[]>(`/api/v1/search?q=${encodeURIComponent(finalQuery)}&size=50`, {
        headers: {
          Authorization: `Bearer ${token}`
        }
      });
      setLogs(response.data || []);
    } catch (err: any) {
      console.error(err);
      if (err.response?.status === 401) {
        handleLogout();
      } else {
        setError('Failed to fetch logs. Is the backend running?');
      }
      setLogs([]);
    } finally {
      setIsLoading(false);
    }
  };

  if (isAuthLoading) {
    return (
      <div className="flex h-screen items-center justify-center bg-background-dark text-white">
        <div className="animate-pulse text-xl">Loading Log Beacon...</div>
      </div>
    );
  }

  if (!token) {
    return <Auth onLogin={handleLogin} hasUsers={hasUsers} />;
  }

  return (
    <div className="relative flex h-screen w-full flex-col overflow-hidden bg-background-dark text-text-light font-display">
      <Header
        query={query}
        setQuery={setQuery}
        onSearch={handleSearch}
        isLiveTail={isLiveTail}
        onToggleLiveTail={toggleLiveTail}
        onLogout={handleLogout}
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
