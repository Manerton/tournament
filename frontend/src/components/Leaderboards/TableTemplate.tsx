import { useEffect, useState, useMemo } from "react";
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
    sortKey2?: keyof T | null;
}

const STORAGE_KEY = "teamColors"; // Ключ для localStorage

// Функция для извлечения hue из hsl строки
function getHue(color: string): number {
    const match = color.match(/hsl\((\d+),\s*\d+%,\s*\d+%\)/);
    return match ? parseInt(match[1], 10) : 0;
}

// Функция для генерации случайного более насыщенного цвета, уникального относительно существующих,
// с минимальным шагом hue в 22.5 градусов (360/16)
function generateUniqueDistinctColor(existingColors: string[]): string {
    const MIN_HUE_STEP = 360 / 16; // 22.5
    const existingHues = existingColors.map(getHue);
    let newColor: string;
    let attempts = 0;
    const maxAttempts = 100; // Чтобы избежать бесконечного цикла

    do {
        const hue = Math.floor(Math.random() * 360);
        const isTooClose = existingHues.some((eh) => {
            const diff = Math.abs(hue - eh);
            const circularDiff = Math.min(diff, 360 - diff);
            return circularDiff < MIN_HUE_STEP;
        });

        if (!isTooClose || attempts >= maxAttempts) {
            const saturation = Math.floor(50 + Math.random() * 50); // 50-80%
            const lightness = Math.floor(60 + Math.random() * 30); // 60-80%
            newColor = `hsl(${hue}, ${saturation}%, ${lightness}%)`;
            // Проверяем на точное совпадение цвета (маловероятно, но на всякий случай)
            if (existingColors.includes(newColor) && attempts < maxAttempts) {
                attempts++;
                continue;
            }
            break;
        }
        attempts++;
    } while (true);

    if (attempts >= maxAttempts) {
        console.warn("Не удалось найти hue с минимальным шагом после максимума попыток. Использую случайный.");
    }

    return newColor;
}

export default function LiveTable<T>({
                                         title,
                                         url,
                                         interval = 1000,
                                         idKey,
                                         columns,
                                         sortKey,
                                         sortKey2 = null,
                                     }: LiveTableProps<T>) {
    const [rows, setRows] = useState<T[]>([]);
    const [colorMap, setColorMap] = useState<Record<string, string>>({});

    // Загрузка цветов из localStorage при монтировании
    useEffect(() => {
        const storedColors = localStorage.getItem(STORAGE_KEY);
        if (storedColors) {
            setColorMap(JSON.parse(storedColors));
        }
    }, []);

    // Сохранение цветов в localStorage при изменении map
    useEffect(() => {
        if (Object.keys(colorMap).length > 0) {
            localStorage.setItem(STORAGE_KEY, JSON.stringify(colorMap));
        }
    }, [colorMap]);

    // Определение ключа для поля команды (по label)
    const teamColumnKey = useMemo(() => {
        const teamCol = columns.find((col) => col.label === "Название команды");
        return teamCol ? teamCol.key : null;
    }, [columns]);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const res = await axios.get(url);
                const data = res.data?.data ?? [];
                if (!Array.isArray(data)) {
                    console.error("Expected 'data' to be an array but got:", data);
                    return;
                }

                // Сортировка (как в оригинале)
                data.sort((a, b) => {
                    const primaryDiff = (b[sortKey] as number) - (a[sortKey] as number);
                    if (primaryDiff !== 0) return primaryDiff;
                    if (sortKey2) {
                        const secondaryDiff = (b[sortKey2] as number) - (a[sortKey2] as number);
                        if (secondaryDiff !== 0) return secondaryDiff;
                    }
                    return String(a[idKey]).localeCompare(String(b[idKey]));
                });

                // Генерация цветов для новых команд (если teamColumnKey есть)
                if (teamColumnKey) {
                    setColorMap((prevMap) => {
                        const newMap = { ...prevMap };
                        const existingColors = Object.values(newMap);
                        data.forEach((row) => {
                            const teamName = String(row[teamColumnKey]);
                            if (!newMap[teamName]) {
                                newMap[teamName] = generateUniqueDistinctColor(existingColors);
                                existingColors.push(newMap[teamName]); // Обновляем список для следующей итерации
                            }
                        });
                        return newMap;
                    });
                }

                setRows(data);
            } catch (err) {
                console.error(`Error fetching data from ${url}:`, err);
            }
        };

        fetchData();
        const timer = setInterval(fetchData, interval);
        return () => clearInterval(timer);
    }, [url, interval, sortKey, sortKey2, teamColumnKey, idKey]);

    // Функция для изменения цвета с проверкой на уникальность
    const handleColorChange = (teamName: string, newColor: string) => {
        const currentColor = colorMap[teamName];
        if (newColor === currentColor) return; // Нет изменений

        const otherColors = Object.entries(colorMap)
            .filter(([key]) => key !== teamName)
            .map(([, value]) => value);

        if (otherColors.includes(newColor)) {
            alert("Этот цвет уже используется другой командой. Выберите другой.");
            return;
        }

        setColorMap((prev) => ({ ...prev, [teamName]: newColor }));
    };

    return (
        <div className="p-1">
            <h2 className="text-xl font-bold mb-2">{title}</h2>
            <hr />
            <table className="table-hover table-bordered w-100 liquid-glass-bg shadow overflow-auto" style={{ borderCollapse: "separate", borderSpacing: "0 10px" }}>
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
                    {rows.map((row, index) => {
                        const teamName = teamColumnKey ? String(row[teamColumnKey]) : "";
                        const teamColor = colorMap[teamName] || "#ffffff"; // Белый по умолчанию

                        // Определение цвета для места (серебро темнее для контраста с #e0e0e0)
                        let placeColor: string | undefined;
                        if (index === 0) {
                            placeColor = "gold";
                        } else if (index === 1) {
                            placeColor = "#a0a0a0"; // Темнее серебра для контраста
                        } else if (index === 2) {
                            placeColor = "#cd7f32"; // Бронза
                        }

                        // Анимация мигания (пульсации) с уменьшенными размерами теней
                        const placeAnimate = placeColor
                            ? {
                                boxShadow: [
                                    `0 0 2px 1px ${placeColor}`,
                                    `0 0 5px 3px ${placeColor}`,
                                ],
                            }
                            : undefined;

                        // Transition с разделением: layout - spring, boxShadow - tween
                        const rowTransition = {
                            layout: { type: "spring", stiffness: 300, damping: 30 },
                        };

                        if (placeColor) {
                            rowTransition.boxShadow = {
                                type: "tween",
                                repeat: Infinity,
                                repeatType: "reverse", // Для пульсации назад без третьего keyframe
                                duration: 1.5,
                                ease: "easeInOut",
                            };
                        }

                        return (
                            <motion.tr
                                key={row[idKey] as string}
                                layout
                                animate={placeAnimate}
                                transition={rowTransition}
                                className="hover:bg-gray-100"
                                // Если хотите раскрасить всю строку: style={{ backgroundColor: teamColor }}
                            >
                                {columns.map((col) => {
                                    const isTeamCell = col.key === teamColumnKey;
                                    return (
                                        <td
                                            key={col.key as string}
                                            className={`border-b p-3 transition-colors duration-300 ${isTeamCell ? 'cursor-pointer' : ''}`}
                                            style={
                                                isTeamCell
                                                    ? { backgroundColor: teamColor }
                                                    : undefined
                                            }
                                            onClick={
                                                isTeamCell
                                                    ? () => {
                                                        // Открываем color-picker при клике
                                                        const input = document.createElement("input");
                                                        input.type = "color";
                                                        input.value = teamColor;
                                                        input.onchange = (e) =>
                                                            handleColorChange(
                                                                teamName,
                                                                (e.target as HTMLInputElement).value
                                                            );
                                                        input.click();
                                                    }
                                                    : undefined
                                            }
                                        >
                                            {row[col.key] as React.ReactNode}
                                        </td>
                                    );
                                })}
                            </motion.tr>
                        );
                    })}
                </AnimatePresence>
                </tbody>
            </table>
        </div>
    );
}