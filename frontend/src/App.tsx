import { useState, useCallback, useEffect } from 'react';
import axios from 'axios';
import './App.css'; // Import our custom styles

// Define the structure of a log entry based on our backend model
interface LogEntry {
  timestamp: string;
  level: string;
  message: string;
  labels: Record<string, string>;
}

const PAGE_SIZE = 50;

function App() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<LogEntry[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [theme, setTheme] = useState(() => localStorage.getItem('theme') || 'light');

  // Apply theme to body and save preference
  useEffect(() => {
    document.body.parentElement?.setAttribute('class', theme === 'dark' ? 'dark-mode' : '');
    localStorage.setItem('theme', theme);
  }, [theme]);

  const handleSearch = useCallback(async (newPage: number) => {
    if (!query) {
      setResults([]);
      return;
    }
    setIsLoading(true);
    setError(null);
    setPage(newPage);

    try {
      const response = await axios.get<LogEntry[]>(`/api/v1/search?q=${query}&page=${newPage}&size=${PAGE_SIZE}`);
      setResults(response.data || []);
    } catch (err) {
      setError('Failed to fetch search results. Is the backend running?');
      console.error(err);
      setResults([]);
    } finally {
      setIsLoading(false);
    }
  }, [query]);

  // Reset page to 1 when a new query is typed
  useEffect(() => {
    setPage(1);
  }, [query]);

  const toggleTheme = () => {
    setTheme(theme === 'light' ? 'dark' : 'light');
  };

  const PaginationControls = () => (
    <div className="d-flex justify-content-between align-items-center my-3">
      <button 
        className="btn btn-secondary" 
        onClick={() => handleSearch(page - 1)} 
        disabled={page <= 1 || isLoading}
      >
        &larr;
      </button>
      <button 
        className="btn btn-secondary" 
        onClick={() => handleSearch(page + 1)} 
        disabled={results.length < PAGE_SIZE || isLoading}
      >
        &rarr;
      </button>
    </div>
  );

  return (
    <div className="app-container">
      <div className="form-check form-switch theme-toggle">
        <input className="form-check-input" type="checkbox" role="switch" id="themeSwitch" checked={theme === 'dark'} onChange={toggleTheme} />
        <label className="form-check-label" htmlFor="themeSwitch">Dark Mode</label>
      </div>

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
                onKeyDown={(e) => e.key === 'Enter' && handleSearch(1)}
              />
              <button className="btn btn-primary px-4" type="button" onClick={() => handleSearch(1)} disabled={isLoading}>
                {isLoading && page === 1 ? (
                  <span className="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>
                ) : (
                  'Search'
                )}
              </button>
            </div>
          </div>
        </div>

        {error && <div className="alert alert-danger mx-auto" style={{ maxWidth: '960px' }}>{error}</div>}

        <div className="results-container">
          <div className="table-responsive">
            <table className={`table table-striped table-hover ${theme === 'dark' ? 'table-dark' : ''}`}>
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
                    <td colSpan={3} className="text-center text-muted py-5">Loading results...</td>
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
                    <td colSpan={3} className="text-center text-muted py-5">
                      No logs found. Enter a query and click search.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
          {results.length > 0 && <PaginationControls />}
        </div>
      </main>
    </div>
  );
}

export default App;
