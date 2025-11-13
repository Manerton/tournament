import React, { useEffect, useState, useCallback, useContext, useRef } from "react";
import { Container, Card, Badge, Spinner, Alert, Row, Col, ProgressBar } from "react-bootstrap";
import { useParams, useNavigate } from "react-router-dom";
import { axiosGetGames } from "../../requests/gamesRequests";
import { GameStatus, type GameAllInformation } from "../../types/gameType";
import "../../assets/GamesPage.css";
import { useAuth } from "../helpers/AuthContext";
import { ThemeContext } from "../helpers/ThemeContext";
import { EventStatus, type EventResponse } from "../../types/TournamentTypes";
import { axiosFinishTour, axiosGetEventById, axiosNextTour } from "../../requests/tournamentRequests";
import { UserRole } from "../../types/userType";
import { API_URL } from "../../config/routes";

const REFRESH_INTERVAL = 1000;
const REFRESH_INTERVAL_TOUR = 5000;

// Ключ для localStorage: уникальный для каждого тура
const getTableNumbersKey = (tourId: string) => `tableNumbers_tour_${tourId}`;

const GamesPage: React.FC = () => {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();

    const { user } = useAuth();
    const alertRef = useRef<HTMLDivElement | null>(null);

    const [tour, setTour] = useState<EventResponse | null>(null);
    const [games, setGames] = useState<GameAllInformation[]>([]);
    const [loading, setLoading] = useState(true);
    const { initialized } = useAuth();
    const { theme } = useContext(ThemeContext);
    const [notify, setNotify] = useState<{ variant: "success" | "danger"; text: string } | null>(null);

    // Загружаем номера столов из localStorage при смене тура
    const [tableAssignments, setTableAssignments] = useState<Record<string, number>>({});

    const startSoundRef = useRef<HTMLAudioElement | null>(null);
    const finishSoundRef = useRef<HTMLAudioElement | null>(null);

    useEffect(() => {
        startSoundRef.current = new Audio("/start.mp3");
        finishSoundRef.current = new Audio("/end.mp3");

        // Опционально: уменьшить громкость, если слишком громко
        if (startSoundRef.current) startSoundRef.current.volume = 0.7;
        if (finishSoundRef.current) finishSoundRef.current.volume = 0.8;
    }, []);

    useEffect(() => {
        if (!id) return;

        const saved = localStorage.getItem(getTableNumbersKey(id));
        if (saved) {
            setTableAssignments(JSON.parse(saved));
        } else {
            setTableAssignments({});
        }
    }, [id]);

    // Сохраняем в localStorage при изменении
    useEffect(() => {
        if (!id || Object.keys(tableAssignments).length === 0) return;
        localStorage.setItem(getTableNumbersKey(id), JSON.stringify(tableAssignments));
    }, [tableAssignments, id]);

    const loadGames = useCallback(async () => {
        try {
            const res = await axiosGetGames(id!);

            setGames((prevGames) => {
                // Определяем, есть ли новые игры
                const newGames = res.filter(g => !prevGames.some(p => p.id === g.id));
                if (newGames.length === 0 && prevGames.length === res.length) {
                    // Ничего не изменилось — оставляем как есть
                    return res;
                }

                // Назначаем номера только новым играм
                setTableAssignments(prev => {
                    const updated = { ...prev };
                    let nextTableNumber = Math.max(...Object.values(prev), 0) + 1;

                    newGames.forEach(game => {
                        if (!updated[game.id]) {
                            updated[game.id] = nextTableNumber++;
                        }
                    });

                    return updated;
                });

                return res;
            });
        } catch (err) {
            console.error("Ошибка загрузки игр:", err);
            const message = err instanceof Error ? err.message : "Не удалось загрузить игры";
            setNotify({ variant: "danger", text: message });
        } finally {
            setLoading(false);
        }
    }, [id]);

    const loadTour = async (silent = true) => {
        try {
            const res = await axiosGetEventById(id!);
            setTour(res);
        } catch (err) {
            if (!silent) {
                const message = err instanceof Error ? err.message : "Не удалось загрузить тур";
                setNotify({ variant: "danger", text: message });
            }
        }
    };

    useEffect(() => {
        loadTour(false); // первый раз с уведомлением
    }, [id]);

    useEffect(() => {
        loadGames();
    }, [loadGames]);

    useEffect(() => {
        if (!id) return;

        const interval = setInterval(() => {
            loadTour(true); // последующие — тихо
        }, REFRESH_INTERVAL_TOUR);

        return () => clearInterval(interval);
    }, [id]);

    useEffect(() => {
        if (!tour || tour.status !== EventStatus.Active) return;

        const interval = setInterval(loadGames, REFRESH_INTERVAL);
        return () => clearInterval(interval);
    }, [loadGames, tour]);

    // Сортировка: активные наверх, завершённые вниз, сохраняя порядок назначения
    const sortedGames = [...games].sort((a, b) => {
        const aFinished = a.game_status === GameStatus.TeamAWon || a.game_status === GameStatus.TeamBWon;
        const bFinished = b.game_status === GameStatus.TeamAWon || b.game_status === GameStatus.TeamBWon;

        if (aFinished && !bFinished) return 1;
        if (!aFinished && bFinished) return -1;

        // Если обе в одном статусе — сохраняем порядок по номеру стола
        const tableA = tableAssignments[a.id] || 999;
        const tableB = tableAssignments[b.id] || 999;
        return tableA - tableB;
    });

    const renderStatus = (status: number) => {
        const isFinished = status === GameStatus.TeamAWon || status === GameStatus.TeamBWon;
        if (isFinished) {
            return <Badge bg="success">Завершена</Badge>;
        }
        return <Badge bg={theme === "light" ? "primary" : "warning"}>В игре</Badge>;
    };

    const handleTeamClick = (game: GameAllInformation) => {
        if (!initialized || game.game_status !== GameStatus.NotFinished) return;
        navigate(`/vote`, { state: { id: game.id, mainEventId: tour?.previous_event_id } });
    };

    const handleFinishTour = async () => {
        try {
            await axiosFinishTour(id!);
            await loadTour();
        } catch (err) {
            setNotify({ variant: "danger", text: "Не удалось завершить тур" });
        }
    };

    const handleNextTour = async () => {
        try {
            const res = await axiosNextTour(id!);
            // Очищаем номера столов для нового тура
            localStorage.removeItem(getTableNumbersKey(id!));
            navigate(`/tours/${res}/games`);
        } catch (err) {
            setNotify({ variant: "danger", text: "Не удалось перейти к следующему туру" });
        }
    };


    const [serverStartTime, setServerStartTime] = useState<number | null>(null); // когда стартанул таймер (unix sec)
    const [timeLeft, setTimeLeft] = useState<number>(300);
    const [ws, setWs] = useState<WebSocket | null>(null);

    // Подключаемся к WS при монтировании
    useEffect(() => {
        if (!id) return;

        // Самый лучший вариант — работает везде:
        const wsUrl = `${API_URL}/api/ws/${id}`.replace(/^http/, 'ws');
        const socket = new WebSocket(wsUrl);

        socket.onopen = () => {
            console.log("[ws] connected to", wsUrl);
            setWs(socket);
        };

        socket.onmessage = (event) => {
            try {
                const msg = JSON.parse(event.data);
                console.log("[ws] received:", msg);

                switch (msg.action) {
                    case "start_timer":
                        const startAt = msg.payload?.start_at;
                        if (typeof startAt === "number") {
                            setServerStartTime(startAt);
                            // Синхронизируем время с учётом задержки сети
                            const elapsed = Math.floor(Date.now() / 1000) - startAt;
                            const remaining = Math.max(300 - elapsed, 0);
                            setTimeLeft(remaining);
                        }
                        break;

                    case "reset_timer":
                        setServerStartTime(null);
                        setTimeLeft(300);
                        break;
                }
            } catch (err) {
                console.error("[ws] parse error:", err);
            }
        };

        socket.onclose = () => {
            console.log("[ws] disconnected");
            setWs(null);
        };

        socket.onerror = (err) => {
            console.error("[ws] error:", err);
            setNotify({ variant: "danger", text: "Ошибка WebSocket" });
        };

        return () => {
            socket.close();
        };
    }, [id]);

    // Синхронный тик каждую секунду (только если таймер запущен сервером)
    useEffect(() => {
        if (serverStartTime === null) return;

        const tick = () => {
            const elapsed = Math.floor(Date.now() / 1000) - serverStartTime;
            const remaining = Math.max(300 - elapsed, 0);
            setTimeLeft(remaining);

            if (remaining <= 0) {
                setServerStartTime(null);
                setNotify({ variant: "success", text: "Время вышло!" });

                finishSoundRef.current?.play().catch(() => {
                console.warn("Звук финиша не воспроизвёлся");
            });
            }
        };

        tick(); // сразу обновить
        const interval = setInterval(tick, 1000);
        return () => clearInterval(interval);
    }, [serverStartTime]);

    // Функции для админа
    const handleStart = () => {
        if (ws?.readyState === WebSocket.OPEN) {
            ws.send("start_timer");

            startSoundRef.current?.play().catch(() => {
            // Игнорируем ошибку, если пользователь ещё не взаимодействовал со страницей
            console.warn("Звук старта не воспроизвёлся — требуется взаимодействие пользователя");
        });
        }
    };

    const handleReset = () => {
        if (ws?.readyState === WebSocket.OPEN) {
            ws.send("reset_timer");
        }
    };

    const progress = (timeLeft / 300) * 100;
    const isRunning = serverStartTime !== null;

    return (
        <Container className="py-4">
            <h2 className="mb-4 fw-bold text-center display-6 d-flex justify-content-center align-items-center gap-4 flex-wrap">
                <span>{tour?.name}</span>

                <div className="d-flex flex-column align-items-center" style={{ width: "200px" }}>
                    <div className="d-flex justify-content-between w-100">
                        <small>{Math.floor(timeLeft / 60)}:{(timeLeft % 60).toString().padStart(2, "0")}</small>
                        <small>{isRunning ? "⏳" : "⏸️"}</small>
                    </div>
                    <ProgressBar
                        now={progress}
                        variant={progress > 60 ? "success" : progress > 30 ? "warning" : "danger"}
                        style={{ height: "8px" }}
                        className="mb-1 w-100"
                    />

                    {/* Кнопки ТОЛЬКО для админа */}
                    {user?.role === UserRole.Admin && (
                        <div className="d-flex justify-content-center gap-2 mt-2">
                            <button
                                className={`btn btn-sm ${isRunning ? "btn-outline-danger" : "btn-outline-success"}`}
                                onClick={handleStart}
                                disabled={!ws || ws.readyState !== WebSocket.OPEN}
                            >
                                {isRunning ? "Идёт..." : "Старт"}
                            </button>
                            <button
                                className="btn btn-sm btn-outline-danger"
                                onClick={handleReset}
                                disabled={!ws || ws.readyState !== WebSocket.OPEN || !isRunning}
                            >
                                Сброс
                            </button>
                        </div>
                    )}
                </div>
            </h2>

            {notify && (
                <div ref={alertRef}>
                    <Alert variant={notify.variant}>{notify.text}</Alert>
                </div>
            )}

            {loading ? (
                <div className="text-center my-5">
                    <Spinner animation="border" />
                </div>
            ) : games.length === 0 ? (
                <Alert variant="info">Нет созданных игр</Alert>
            ) : (
                <>
                    <Row xs={1} md={2} className="g-4">
                        {sortedGames.map((g) => (
                            <Col key={g.id}>
                                <Card
                                    className={`rounded-3 shadow-sm game-card liquid-glass-bg ${g.game_status === GameStatus.TeamAWon || g.game_status === GameStatus.TeamBWon ? 'opacity-75' : ''}`}
                                    onClick={() => handleTeamClick(g)}
                                    style={{ cursor: initialized && g.game_status === GameStatus.NotFinished ? 'pointer' : 'default' }}
                                >
                                    <Card.Header className="text-center fw-bold">
                                        Стол №{tableAssignments[g.id] || "?"}
                                    </Card.Header>

                                    <Card.Body className="d-flex align-items-center justify-content-between text-center">
                                        <div className={`team-name ${g.game_status === GameStatus.TeamAWon ? "winner" : g.game_status === GameStatus.TeamBWon ? "loser" : ""}`}>
                                            <h5>{g.team_a_name}</h5>
                                            <div className="score">{g.team_a_score}</div>
                                        </div>

                                        <div className="vs-section">
                                            <div className="vs-text">VS</div>
                                            {renderStatus(g.game_status)}
                                        </div>

                                        <div className={`team-name ${g.game_status === GameStatus.TeamBWon ? "winner" : g.game_status === GameStatus.TeamAWon ? "loser" : ""}`}>
                                            <h5>{g.team_b_name}</h5>
                                            <div className="score">{g.team_b_score}</div>
                                        </div>
                                    </Card.Body>
                                </Card>
                            </Col>
                        ))}
                    </Row>

                    <div className="text-center mt-5">
                        {tour?.status === EventStatus.Active && user?.role === UserRole.Admin ? (
                            <button className="btn btn-lg btn-outline-danger" onClick={handleFinishTour}>
                                Завершить тур
                            </button>
                        ) : tour && tour.stage >= tour.max_stage ? (
                            <Alert variant="info" className="d-inline-block">
                                Достигнуто максимальное количество туров
                            </Alert>
                        ) : tour?.status === EventStatus.Closed ? (
                            <button className="btn btn-lg btn-outline-primary" onClick={handleNextTour}>
                                Перейти к следующему туру
                            </button>
                        ) : null}
                    </div>
                </>
            )}
        </Container>
    );
};

export default GamesPage;