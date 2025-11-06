import React, { useEffect, useState, useCallback, useContext, useRef } from "react";
import { Container, Card, Badge, Spinner, Alert, Row, Col } from "react-bootstrap";
import { useParams, useNavigate } from "react-router-dom";
import { axiosGetGames } from "../../requests/gamesRequests";
import type { GameAllInformation } from "../../types/gameType";
import "../../assets/GamesPage.css";
import { useAuth } from "../helpers/AuthContext";
import { ThemeContext } from "../helpers/ThemeContext";
import { EventStatus, type EventResponse } from "../../types/TournamentTypes";
import { axiosFinishTour, axiosGetEventById, axiosNextTour } from "../../requests/tournamentRequests";

const REFRESH_INTERVAL = 1000;

const GamesPage: React.FC = () => {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const alertRef = useRef<HTMLDivElement | null>(null);

    const [tour, setTour] = useState<EventResponse>();
    const [games, setGames] = useState<GameAllInformation[]>([]);
    const [loading, setLoading] = useState(true);
    const { initialized } = useAuth();
    const { theme } = useContext(ThemeContext);
    const [notify, setNotify] = useState<{ variant: "success" | "danger"; text: string } | null>(null);

    const loadGames = useCallback(async () => {
        try {
            const res = await axiosGetGames(id!);
            setGames((prev) => (JSON.stringify(prev) !== JSON.stringify(res) ? res : prev));
        } catch (err) {
            console.error("Ошибка загрузки игр:", err);
            const message = err instanceof Error ? err.message : "Не удалось загрузить игры";
            setNotify({ variant: "danger", text: message });
        } finally {
            setLoading(false);
        }
    }, [id]);

    const loadTour = async () => {
        try {
            const res = await axiosGetEventById(id!);
            setTour(res);
        } catch (err) {
            const message = err instanceof Error ? err.message : "Не удалось загрузить тур";
            setNotify({ variant: "danger", text: message });
        }
    };

    useEffect(() => {
        loadTour();
    }, [id]);

    useEffect(() => {
        loadGames();
    }, [loadGames]);

    useEffect(() => {
        const interval = setInterval(loadGames, REFRESH_INTERVAL);
        return () => clearInterval(interval);
    }, [loadGames]);

    const renderStatus = (status: number) => {
        switch (status) {
            case 3:
                return (
                    <Badge bg={theme === "light" ? "primary" : "warning"}>В игре</Badge>
                );
            default:
                return <Badge bg="success">Завершена</Badge>;
        }
    };

    const handleTeamClick = (game: GameAllInformation, mainEventId: string) => {
        if (!initialized) return;
        navigate(`/vote`, { state: { game, mainEventId } });
    };

    const handleFinishTour = async () => {
        try {
            await axiosFinishTour(id!);
            await loadTour();
        } catch (err) {
            const message = err instanceof Error ? err.message : "Не удалось завершить тур";
            setNotify({ variant: "danger", text: message });
        }
    };

    const handleNextTour = async () => {
        try {
            const res = await axiosNextTour(id!);
            navigate(`/tours/${res}/games`);
        } catch (err) {
            const message = err instanceof Error ? err.message : "Не удалось перейти к следующему туру";
            setNotify({ variant: "danger", text: message });
        }
    };

    return (
        <Container className="py-4">
            <h2 className="mb-4 fw-bold text-center display-6">
                Игры {tour?.name}
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
                        {games.map((g) => (
                            <Col key={g.team_a_id + g.team_b_id}>
                                <Card
                                    className="rounded-3 shadow-sm game-card liquid-glass-bg "
                                    onClick={() => handleTeamClick(g, tour?.previous_event_id!)}
                                >
                                    <Card.Body className="d-flex align-items-center justify-content-between text-center">
                                        <div
                                            className={`team-name  ${g.game_status !== 3 && g.team_a_score > g.team_b_score
                                                    ? "winner"
                                                    : g.game_status !== 3 && g.team_a_score < g.team_b_score
                                                        ? "loser"
                                                        : ""
                                                }`}
                                        >
                                            <h5>{g.team_a_name}</h5>
                                            <div className="score">{g.team_a_score}</div>
                                        </div>

                                        <div className="vs-section">
                                            <div className="vs-text">VS</div>
                                            {renderStatus(g.game_status)}
                                        </div>

                                        <div
                                            className={`team-name ${g.game_status !== 3 && g.team_b_score > g.team_a_score
                                                    ? "winner"
                                                    : g.game_status !== 3 && g.team_b_score < g.team_a_score
                                                        ? "loser"
                                                        : ""
                                                }`}
                                        >
                                            <h5>{g.team_b_name}</h5>
                                            <div className="score">{g.team_b_score}</div>
                                        </div>
                                    </Card.Body>
                                </Card>
                            </Col>
                        ))}
                    </Row>

                    <div className="text-center mt-4">
                        {tour?.status === EventStatus.Active ? (
                            <button
                                className="btn btn-outline-danger"
                                onClick={handleFinishTour}
                            >
                                Завершить тур
                            </button>
                        ) : tour && tour.stage >= tour.max_stage ? (
                            <Alert variant="info" className="d-inline-block">
                                🏁 Достигнуто максимальное количество туров
                            </Alert>
                        ) : (
                            <button
                                className="btn btn-outline-primary"
                                onClick={handleNextTour}
                            >
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
