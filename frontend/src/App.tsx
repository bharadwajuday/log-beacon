import { useState, useCallback } from 'react';
import axios from 'axios';
import './App.css'; // Import our custom styles

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
    <div className="app-container">
      <header className="text-center mb-4">
        <h1>Log Beacon</h1>
        <p className="lead text-muted">Your centralized log search</p>
      </header>

      <main>
        <div className="search-section">
          <div className="search-container">
            <div className="input-group input-group-lg mb-3">
              <input
                type="text"
                className="form-control"
                placeholder="Search logs..."
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
              />
                          <button className="btn btn-primary" type="button" onClick={handleSearch} disabled={isLoading}>
                            {isLoading ? (
                              <span className="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>
                            ) : (
                              'Search'
                            )}
                          </button>            </div>
          </div>
        </div>

        {error && <div className="alert alert-danger mx-auto" style={{ maxWidth: '960px' }}>{error}</div>}

        <div className="results-container">
          <div className="table-responsive">
            <table className="table table-striped table-hover">
              <thead className="table-light">
                <tr>
                  <th scope="col">Timestamp</th>
                  <th scope="col">Level</th>
                  <th scope="col">Message</th>
                </tr>
              </thead>
              <tbody>
                {isLoading ? (
                  <tr>
                    <td colSpan={3} className="text-center text-muted">Loading results...</td>
                  </tr>
                ) : error ? (
                  <tr>
                    <td colSpan={3} className="text-center text-danger">{error}</td>
                  </tr>
                ) : results.length > 0 ? (
                  results.map((log, index) => (
                    <tr key={index}>
                      <td className="text-nowrap">{new Date(log.timestamp).toLocaleString()}</td>
                      <td>
                        <span className={`badge bg-${log.level === 'error' ? 'danger' : 'secondary'}`}>
                          {log.level || 'info'}
                        </span>
                      </td>
                      <td>{log.message}</td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td colSpan={3} className="text-center text-muted">
                      No logs found. Enter a query and click search.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      </main>
    </div>
  );
}

export default App;