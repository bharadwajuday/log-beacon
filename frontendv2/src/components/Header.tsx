import React from 'react';

interface HeaderProps {
    query: string;
    setQuery: (query: string) => void;
    onSearch: () => void;
    isLiveTail: boolean;
    onToggleLiveTail: () => void;
    onLogout: () => void;
}

const Header: React.FC<HeaderProps> = ({ query, setQuery, onSearch, isLiveTail, onToggleLiveTail, onLogout }) => {
    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter' && !isLiveTail) {
            onSearch();
        }
    };

    return (
        <header className="sticky top-0 z-10 flex items-center justify-between whitespace-nowrap border-b border-solid border-border-dark bg-background-dark px-6 py-3">
            <div className="flex items-center gap-8">
                <div className="flex items-center gap-4 text-text-light">
                    <div className="size-6 text-primary">
                        <svg fill="none" viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg" className="w-6 h-6">
                            <path d="M44 11.2727C44 14.0109 39.8386 16.3957 33.69 17.6364C39.8386 18.877 44 21.2618 44 24C44 26.7382 39.8386 29.123 33.69 30.3636C39.8386 31.6043 44 33.9891 44 36.7273C44 40.7439 35.0457 44 24 44C12.9543 44 4 40.7439 4 36.7273C4 33.9891 8.16144 31.6043 14.31 30.3636C8.16144 29.123 4 26.7382 4 24C4 21.2618 8.16144 18.877 14.31 17.6364C8.16144 16.3957 4 14.0109 4 11.2727C4 7.25611 12.9543 4 24 4C35.0457 4 44 7.25611 44 11.2727Z" fill="currentColor"></path>
                        </svg>
                    </div>
                    <h2 className="text-text-light text-lg font-bold leading-tight tracking-[-0.015em]">Log Beacon</h2>
                </div>
            </div>
            <div className="flex flex-1 justify-center px-8">
                <label className="flex flex-col w-full !h-10 max-w-2xl">
                    <div className="flex w-full flex-1 items-stretch rounded-lg h-full">
                        <input
                            className={`form-input flex w-full min-w-0 flex-1 resize-none overflow-hidden rounded-l-lg text-text-light focus:outline-0 focus:ring-0 border-y border-l border-border-dark bg-panel-dark focus:border-primary h-full placeholder:text-text-subtle-dark px-4 text-base font-normal leading-normal ${isLiveTail ? 'opacity-50 cursor-not-allowed' : ''}`}
                            placeholder={isLiveTail ? "Live Tail Active..." : "Search logs..."}
                            value={query}
                            onChange={(e) => setQuery(e.target.value)}
                            onKeyDown={handleKeyDown}
                            disabled={isLiveTail}
                        />
                        <button
                            className={`flex min-w-[84px] max-w-[480px] cursor-pointer items-center justify-center overflow-hidden rounded-r-lg h-10 px-4 text-background-dark text-sm font-bold leading-normal tracking-[0.015em] ${isLiveTail ? 'bg-gray-500 cursor-not-allowed' : 'bg-primary'}`}
                            onClick={onSearch}
                            disabled={isLiveTail}
                        >
                            <span className="truncate">Search</span>
                        </button>
                    </div>
                </label>
            </div>
            <div className="flex flex-initial items-center justify-end gap-2">
                <button
                    className={`flex max-w-[480px] cursor-pointer items-center justify-center overflow-hidden rounded-lg h-10 gap-2 text-sm font-bold leading-normal tracking-[0.015em] min-w-0 px-4 ${isLiveTail ? 'bg-green-600 text-white animate-pulse' : 'bg-panel-dark text-text-light'}`}
                    onClick={onToggleLiveTail}
                >
                    <span className="material-symbols-outlined">{isLiveTail ? 'stop' : 'play_arrow'}</span>
                    <span className="truncate">{isLiveTail ? 'Live' : 'Live Tail'}</span>
                </button>
                <button className="flex max-w-[480px] cursor-pointer items-center justify-center overflow-hidden rounded-lg h-10 bg-panel-dark text-text-light gap-2 text-sm font-bold leading-normal tracking-[0.015em] min-w-0 px-2.5">
                    <span className="material-symbols-outlined text-text-light">notifications</span>
                </button>
                <button 
                  onClick={onLogout}
                  className="flex max-w-[480px] cursor-pointer items-center justify-center overflow-hidden rounded-lg h-10 bg-red-500/10 text-red-500 hover:bg-red-500/20 gap-2 text-sm font-bold leading-normal tracking-[0.015em] min-w-0 px-4 transition-colors">
                    <span className="material-symbols-outlined">logout</span>
                    <span className="truncate">Logout</span>
                </button>
            </div>
        </header>
    );
};

export default Header;
