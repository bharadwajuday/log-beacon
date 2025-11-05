import { useState, useCallback } from 'react';
import axios from 'axios';

// Define the structure of a log entry based on our backend model
interface LogEntry {
  timestamp: string;
  level: string;
  message: string;
  labels: Record<string, string>;
}

function App() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<LogEntry[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSearch = useCallback(async () => {
    if (!query) {
      setResults([]);
      return;
    }
    setIsLoading(true);
    setError(null);
    try {
      // Make the API call to our backend. The /api prefix will be proxied by Nginx.
      const response = await axios.get<LogEntry[]>(`/api/v1/search?q=${query}`);
      setResults(response.data || []);
    } catch (err) {
      setError('Failed to fetch search results. Is the backend running?');
      console.error(err);
      setResults([]);
    } finally {
      setIsLoading(false);
    }
  }, [query]);

  return (
    <div className="container mt-4">
      <header className="text-center mb-4">
        <h1>Log Beacon</h1>
        <p className="lead">Search and explore your logs</p>
      </header>

      <div className="input-group mb-3">
        <input
          type="text"
          className="form-control"
          placeholder="Search logs... (e.g., error, authentication, payment)"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
        />
        <button className="btn btn-primary" type="button" onClick={handleSearch} disabled={isLoading}>
          {isLoading ? 'Searching...' : 'Search'}
        </button>
      </div>

      {error && <div className="alert alert-danger">{error}</div>}

      <div className="table-responsive">
        <table className="table table-striped table-hover">
          <thead className="table-dark">
            <tr>
              <th scope="col">Timestamp</th>
              <th scope="col">Level</th>
              <th scope="col">Message</th>
            </tr>
          </thead>
          <tbody>
            {results.length > 0 ? (
              results.map((log, index) => (
                <tr key={index}>
                  <td>{new Date(log.timestamp).toLocaleString()}</td>
                  <td><span className={`badge bg-${log.level === 'error' ? 'danger' : 'secondary'}`}>{log.level || 'info'}</span></td>
                  <td>{log.message}</td>
                </tr>
              ))
            ) : (
              <tr>
                <td colSpan={3} className="text-center text-muted">
                  {isLoading ? 'Loading results...' : 'No logs found. Try a new search.'}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export default App;