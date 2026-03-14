import React, { useState } from 'react';
import axios from 'axios';

interface AuthProps {
  onLogin: (token: string) => void;
  hasUsers: boolean;
}

const Auth: React.FC<AuthProps> = ({ onLogin, hasUsers }) => {
  const [isRegistering, setIsRegistering] = useState(!hasUsers);
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError(null);

    const endpoint = isRegistering ? '/api/v1/auth/register' : '/api/v1/auth/login';

    try {
      const response = await axios.post(endpoint, { username, password });
      
      if (isRegistering) {
        // After successful registration, switch to login mode
        setIsRegistering(false);
        setError('Registration successful! Please login.');
      } else {
        // Successful login
        onLogin(response.data.token);
      }
    } catch (err: any) {
      setError(err.response?.data?.error || 'An error occurred. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-[#0a0a0c] px-4 font-display relative overflow-hidden">
      {/* Decorative background elements */}
      <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-primary-blue/20 blur-[120px] rounded-full animate-pulse"></div>
      <div className="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-purple-600/10 blur-[120px] rounded-full"></div>

      <div className="w-full max-w-md space-y-8 rounded-3xl bg-white/[0.03] p-10 shadow-2xl border border-white/10 backdrop-blur-2xl relative z-10">
        <div className="text-center">
          <div className="inline-flex items-center justify-center p-3 rounded-2xl bg-primary-blue mb-6 shadow-lg shadow-primary-blue/30">
            <svg fill="none" viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg" className="w-8 h-8 text-white">
              <path d="M44 11.2727C44 14.0109 39.8386 16.3957 33.69 17.6364C39.8386 18.877 44 21.2618 44 24C44 26.7382 39.8386 29.123 33.69 30.3636C39.8386 31.6043 44 33.9891 44 36.7273C44 40.7439 35.0457 44 24 44C12.9543 44 4 40.7439 4 36.7273C4 33.9891 8.16144 31.6043 14.31 30.3636C8.16144 29.123 4 26.7382 4 24C4 21.2618 8.16144 18.877 14.31 17.6364C8.16144 16.3957 4 14.0109 4 11.2727C4 7.25611 12.9543 4 24 4C35.0457 4 44 7.25611 44 11.2727Z" fill="currentColor"></path>
            </svg>
          </div>
          <h2 className="text-4xl font-extrabold tracking-tight text-primary mb-2 uppercase italic">
            Log Beacon
          </h2>
          <p className="text-text-muted/60 text-sm">
            {isRegistering ? 'Create an account to get started' : 'Sign in to access your logs'}
          </p>
        </div>

        <form className="mt-10 space-y-6" onSubmit={handleSubmit}>
          <div className="space-y-5">
            <div className="flex flex-col gap-1.5">
              <label className="text-[20px] uppercase tracking-[0.1em] font-bold text-text-muted/50 ml-1">
                Username
              </label>
              <input
                type="text"
                required
                autoComplete="username"
                className="block w-full rounded-2xl border-0 bg-white/[0.04] py-3.5 px-4 text-white shadow-sm ring-1 ring-inset ring-white/10 placeholder:text-text-muted/20 focus:ring-2 focus:ring-inset focus:ring-primary-blue sm:text-sm sm:leading-6 transition-all outline-none hover:bg-white/[0.06]"
                placeholder="Enter your username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <label className="text-[20px] uppercase tracking-[0.1em] font-bold text-text-muted/50 ml-1">
                Password
              </label>
              <input
                type="password"
                required
                autoComplete="current-password"
                className="block w-full rounded-2xl border-0 bg-white/[0.04] py-3.5 px-4 text-white shadow-sm ring-1 ring-inset ring-white/10 placeholder:text-text-muted/20 focus:ring-2 focus:ring-inset focus:ring-primary-blue sm:text-sm sm:leading-6 transition-all outline-none hover:bg-white/[0.06]"
                placeholder="Enter your password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>
          </div>

          {error && (
            <div className={`rounded-xl p-3.5 text-xs font-medium border ${error.includes('successful') ? 'bg-green-500/5 border-green-500/20 text-green-400' : 'bg-red-500/5 border-red-500/20 text-red-400'}`}>
              <div className="flex items-center gap-2">
                <span className="material-symbols-outlined text-sm">{error.includes('successful') ? 'check_circle' : 'error'}</span>
                {error}
              </div>
            </div>
          )}

          <div className="pt-2">
            <button
              type="submit"
              disabled={isLoading}
              className="group relative flex w-full justify-center rounded-2xl bg-gradient-to-r from-primary-blue to-blue-600 px-4 py-3.5 text-sm font-bold text-white shadow-xl shadow-primary-blue/20 hover:shadow-primary-blue/40 transition-all active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed overflow-hidden"
            >
              <div className="absolute inset-0 bg-white/20 translate-y-full group-hover:translate-y-0 transition-transform duration-300"></div>
              <span className="relative">{isLoading ? 'Processing...' : (isRegistering ? 'Create Account' : 'Sign In')}</span>
            </button>
          </div>
        </form>

        <div className="text-center pt-2">
          <button
            onClick={() => setIsRegistering(!isRegistering)}
            className="text-xs font-bold text-text-muted/40 hover:text-primary-blue transition-colors uppercase tracking-wider"
          >
            {isRegistering ? 'Already have an account? Sign in' : 'Don\'t have an account? Register'}
          </button>
        </div>
      </div>
    </div>
  );
};

export default Auth;
