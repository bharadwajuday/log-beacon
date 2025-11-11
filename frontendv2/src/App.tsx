import { useState } from "react";
import axios from "axios";
import Header from "./components/Header";
import Sidebar from "./components/Sidebar";
import LogView from "./components/LogView";
import { LogEntryProps } from "./components/LogEntry";

function App() {
  const [logs, setLogs] = useState<LogEntryProps[]>([]);

  const handleSearch = async (query: string) => {
    try {
      const response = await axios.get(`/api/v1/search?query=${query}`);
      setLogs(response.data);
    } catch (error) {
      console.error("Error fetching logs:", error);
    }
  };

  return (
    <div className="relative flex h-auto min-h-screen w-full flex-col">
      <Header onSearch={handleSearch} />
      <div className="flex flex-1">
        <Sidebar />
        <LogView logs={logs} />
      </div>
    </div>
  )
}

export default App
