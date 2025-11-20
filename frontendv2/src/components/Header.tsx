import React from 'react';

interface HeaderProps {
    query: string;
    setQuery: (query: string) => void;
    onSearch: () => void;
}

const Header: React.FC<HeaderProps> = ({ query, setQuery, onSearch }) => {
    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter') {
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
                            className="form-input flex w-full min-w-0 flex-1 resize-none overflow-hidden rounded-l-lg text-text-light focus:outline-0 focus:ring-0 border-y border-l border-border-dark bg-panel-dark focus:border-primary h-full placeholder:text-text-subtle-dark px-4 text-base font-normal leading-normal"
                            placeholder="Search logs..."
                            value={query}
                            onChange={(e) => setQuery(e.target.value)}
                            onKeyDown={handleKeyDown}
                        />
                        <button
                            className="flex min-w-[84px] max-w-[480px] cursor-pointer items-center justify-center overflow-hidden rounded-r-lg h-10 px-4 bg-primary text-background-dark text-sm font-bold leading-normal tracking-[0.015em]"
                            onClick={onSearch}
                        >
                            <span className="truncate">Search</span>
                        </button>
                    </div>
                </label>
            </div>
            <div className="flex flex-initial items-center justify-end gap-2">
                <button className="flex max-w-[480px] cursor-pointer items-center justify-center overflow-hidden rounded-lg h-10 bg-panel-dark text-text-light gap-2 text-sm font-bold leading-normal tracking-[0.015em] min-w-0 px-2.5">
                    <span className="material-symbols-outlined text-text-light">notifications</span>
                </button>
                <button className="flex max-w-[480px] cursor-pointer items-center justify-center overflow-hidden rounded-lg h-10 bg-panel-dark text-text-light gap-2 text-sm font-bold leading-normal tracking-[0.015em] min-w-0 px-2.5">
                    <span className="material-symbols-outlined text-text-light">help</span>
                </button>
                <div className="bg-center bg-no-repeat aspect-square bg-cover rounded-full size-10" style={{ backgroundImage: 'url("https://lh3.googleusercontent.com/aida-public/AB6AXuDnGJDXzef2v6TZrkKLzSPbNGVwz_mdz7ELiq3XOwV6ET9kCXIOWnvcCvAEIaeA-d7f6gvbxDY65Rb-7n83R_NfNN6tYIZeiUe_7XywydrEBbd2yo0R_5zl9mbqRoO00_LRbC9PMMLYyArs87iZJG7qBjjjT-oE-zG1hzIY_u3CVqAWVmZaQu0-Di16i_DJvuOSAX5SHfFDIw76NvWatlso753EhWejFs4zHpEbUBXKeHWqDl4DOBq6vt8lme-P_Evg69I1TBUjeJLb")' }}></div>
            </div>
        </header>
    );
};

export default Header;
