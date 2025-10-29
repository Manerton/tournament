import { useEffect, useState } from "react";
import axios from "axios";
import { motion, AnimatePresence } from "framer-motion";

interface TableRow {
  id: string; // уникальный ключ, может быть name или id
  name: string;
  score: number;
}

interface LiveTableProps {
  title: string;
  url: string;
  interval?: number; // ms
}

export default function LiveTable({ title, url, interval = 1000 }: LiveTableProps) {
  const [rows, setRows] = useState<TableRow[]>([]);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const res = await axios.get<TableRow[]>(url);
        const data = res.data;
        data.sort((a, b) => b.score - a.score);
        setRows(data);
      } catch (err) {
        console.error(`Error fetching data from ${url}:`, err);
      }
    };

    fetchData();
    const timer = setInterval(fetchData, interval);

    return () => clearInterval(timer);
  }, [url, interval]);

  return (
    <div className="p-4">
      <h2 className="text-xl font-bold mb-2">{title}</h2>
      <div className="space-y-2">
        <AnimatePresence>
          {rows.map((row) => (
            <motion.div
              key={row.id}
              layout
              transition={{ type: "spring", stiffness: 300, damping: 30 }}
              className="bg-white rounded-lg shadow p-3 flex justify-between"
            >
              <span>{row.name}</span>
              <span>{row.score}</span>
            </motion.div>
          ))}
        </AnimatePresence>
      </div>
    </div>
  );
}
