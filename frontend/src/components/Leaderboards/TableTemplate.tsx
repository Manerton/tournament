import { useEffect, useState } from "react";
import axios from "axios";
import { motion, AnimatePresence } from "framer-motion";

interface Column<T> {
    key: keyof T;
    label: string;
}

interface LiveTableProps<T> {
    title: string;
    url: string;
    interval?: number; // ms
    idKey: keyof T;
    columns: Column<T>[];
    sortKey: keyof T;
}

export default function LiveTable<T>({
                                         title,
                                         url,
                                         interval = 1000,
                                         idKey,
                                         columns,
                                         sortKey,
                                         sortKey2 = null,
                                     }: LiveTableProps<T> & { sortKey2?: keyof T | null }) {
    const [rows, setRows] = useState<T[]>([]);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const res = await axios.get(url);
                const data = res.data?.data ?? [];

                if (!Array.isArray(data)) {
                    console.error("Expected 'data' to be an array but got:", data);
                    return;
                }

                data.sort((a, b) => {
                    // 1️⃣ Сначала сортируем по основному ключу
                    const primaryDiff = (b[sortKey] as number) - (a[sortKey] as number);
                    if (primaryDiff !== 0) return primaryDiff;

                    // 2 Если есть sortKey2 — сортируем по нему
                    if (sortKey2) {
                        const secondaryDiff = (b[sortKey2] as number) - (a[sortKey2] as number);
                        if (secondaryDiff !== 0) return secondaryDiff;
                    }

                    // 3Если значения равны — сортируем по idKey для стабильности
                    return String(a[idKey]).localeCompare(String(b[idKey]));
                });

                setRows(data);
            } catch (err) {
                console.error(`Error fetching data from ${url}:`, err);
            }
        };

        fetchData();
        const timer = setInterval(fetchData, interval);
        return () => clearInterval(timer);
    }, [url, interval, sortKey, sortKey2]);

    return (
        <div className="p-1">
            <h2 className="text-xl font-bold mb-2">{title}</h2>
            <hr />
            <table className="table-hover table-bordered w-100 liquid-glass-bg shadow overflow-auto">
                <thead>
                <tr>
                    {columns.map((col) => (
                        <th
                            key={col.key as string}
                            className="border-b p-3 text-left font-semibold"
                        >
                            {col.label}
                        </th>
                    ))}
                </tr>
                </thead>
                <tbody>
                <AnimatePresence>
                    {rows.map((row) => (
                        <motion.tr
                            key={row[idKey] as string}
                            layout
                            transition={{ type: "spring", stiffness: 300, damping: 30 }}
                            className="hover:bg-gray-100"
                        >
                            {columns.map((col) => (
                                <td key={col.key as string} className="border-b p-3">
                                    {row[col.key] as React.ReactNode}
                                </td>
                            ))}
                        </motion.tr>
                    ))}
                </AnimatePresence>
                </tbody>
            </table>
        </div>
    );
}
