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

const REFRESH_INTERVAL = 1000;
const REFRESH_INTERVAL_TOUR = 5000;

// Ключ для localStorage: уникальный для каждого тура
const getTableNumbersKey = (tourId: string) => `tableNumbers_tour_${tourId}`;

const GamesPage: React.FC = () => {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const alertRef = useRef<HTMLDivElement | null>(null);

    const [tour, setTour] = useState<EventResponse | null>(null);
    const [games, setGames] = useState<GameAllInformation[]>([]);
    const [loading, setLoading] = useState(true);
    const { initialized } = useAuth();
    const { theme } = useContext(ThemeContext);
    const [notify, setNotify] = useState<{ variant: "success" | "danger"; text: string } | null>(null);

    // Загружаем номера столов из localStorage при смене тура
    const [tableAssignments, setTableAssignments] = useState<Record<string, number>>({});

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

    // Таймер (оставляем как есть)
    const [timeLeft, setTimeLeft] = useState(300);
    const [isRunning, setIsRunning] = useState(false);

    const toggleTimer = () => setIsRunning(prev => !prev);
    const resetTimer = () => {
        setIsRunning(false);
        setTimeLeft(300);
    };

    useEffect(() => {
        if (!isRunning) return;

        const interval = setInterval(() => {
            setTimeLeft(prev => {
                if (prev <= 1) {
                    setIsRunning(false);
                    setNotify({ variant: "success", text: "Время вышло!" });
                    return 0;
                }
                return prev - 1;
            });
        }, 1000);

        return () => clearInterval(interval);
    }, [isRunning]);

    const progress = (timeLeft / 300) * 100;

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
                    <div className="d-flex justify-content-center gap-2">
                        <button className={`btn btn-sm ${isRunning ? "btn-outline-danger" : "btn-outline-success"}`} onClick={toggleTimer}>
                            {isRunning ? "Пауза" : "Старт"}
                        </button>
                        <button className="btn btn-sm btn-outline-danger" onClick={resetTimer}>
                            Сброс
                        </button>
                    </div>
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
                        {tour?.status === EventStatus.Active ? (
                            <button className="btn btn-lg btn-outline-danger" onClick={handleFinishTour}>
                                Завершить тур
                            </button>
                        ) : tour && tour.stage >= tour.max_stage ? (
                            <Alert variant="info" className="d-inline-block">
                                Достигнуто максимальное количество туров
                            </Alert>
                        ) : (
                            <button className="btn btn-lg btn-outline-primary" onClick={handleNextTour}>
                                Перейти к следующему туру
                            </button>
                        )}
                    </div>
                </>
            )}
        </Container>
    );
};

export default GamesPage;