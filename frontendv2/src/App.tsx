import { useState } from 'react';
import axios from 'axios';
import Header from './components/Header';
import Sidebar from './components/Sidebar';
import LogList from './components/LogList';
import { type LogEntry } from './types';

function App() {
  const [query, setQuery] = useState('');
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [hasSearched, setHasSearched] = useState(false);

  const handleSearch = async () => {
    if (!query.trim()) {
      setLogs([]);
      setError(null);
      return;
    }

    setHasSearched(true);
    setIsLoading(true);
    setError(null);
    try {
      // Use the same API endpoint as the original frontend
      const response = await axios.get<LogEntry[]>(`/api/v1/search?q=${query}&size=50`);
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
      <Header query={query} setQuery={setQuery} onSearch={handleSearch} />
      <div className="flex flex-1 overflow-hidden">
        <Sidebar />
        <LogList logs={logs} isLoading={isLoading} error={error} hasSearched={hasSearched} />
      </div>
    </div>
  );
}

export default App;
