import React, { useEffect, useState, useCallback } from "react";
import { Container, Card, Badge, Spinner, Alert } from "react-bootstrap";
import { useParams, useNavigate } from "react-router-dom";
import { axiosGetGames } from "../../requests/gamesRequests";
import type { GameAllInformation } from "../../types/gameType";
import "../../assets/GamesPage.css";
import { useAuth } from "../helpers/AuthContext";

const REFRESH_INTERVAL = 1000; // каждые 5 секунд обновляем данные

const GamesPage: React.FC = () => {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();

    const [games, setGames] = useState<GameAllInformation[]>([]);
    const [loading, setLoading] = useState(true);

    const {initialized} = useAuth();

    const loadGames = useCallback(async () => {
        try {
            const res = await axiosGetGames(id!);
            setGames((prevGames) => {
                // Оптимизация: обновляем только если данные изменились
                const changed = JSON.stringify(prevGames) !== JSON.stringify(res);
                return changed ? res : prevGames;
            });
        } catch (err) {
            console.error("Ошибка загрузки игр:", err);
        } finally {
            setLoading(false);
        }
    }, [id]);

    // Первичная загрузка
    useEffect(() => {
        loadGames();
    }, [loadGames]);

    // Периодическое обновление каждые N мс
    useEffect(() => {
        const interval = setInterval(loadGames, REFRESH_INTERVAL);
        return () => clearInterval(interval);
    }, [loadGames]);

    const renderStatus = (status: number) => {
        switch (status) {
            case 0:
                return <Badge bg="secondary">Ожидает</Badge>;
            case 1:
                return <Badge bg="warning">В игре</Badge>;
            case 2:
                return <Badge bg="success">Завершена</Badge>;
            default:
                return <Badge bg="light">Неизвестно</Badge>;
        }
    };

    const handleTeamClick = (teamId: string) => {
        if (!initialized) return; // если не авторизован, ничего не делаем
        navigate(`/vote/${teamId}`);
    };

    return (
        <Container className="py-4">
            <h2 className="mb-4 fw-bold text-center display-6">Игры турнира</h2>

            {loading ? (
                <div className="text-center my-5">
                    <Spinner animation="border" />
                </div>
            ) : games.length === 0 ? (
                <Alert variant="info">Нет созданных игр</Alert>
            ) : (
                games.map((g) => (
                    <Card
                        key={g.team_a_id + g.team_b_id}
                        className="mb-4 rounded-3 shadow-sm liquid-glass-bg"
                    >
                        <Card.Body className="d-flex justify-content-between align-items-center py-3 px-2 flex-wrap">
                            {/* Команда A */}
                            <div
                                onClick={() => handleTeamClick(g.team_a_id)}
                                className={`team-card text-center flex-fill me-2 clickable ${
                                    g.game_status === 2 && g.team_a_score > g.team_b_score
                                        ? "winner"
                                        : g.game_status === 2 && g.team_a_score < g.team_b_score
                                        ? "loser"
                                        : ""
                                }`}
                            >
                                <h5>{g.team_a_name}</h5>
                                <div className="score">{g.team_a_score}</div>
                            </div>

                            {/* Счёт и VS */}
                            <div className="vs-container text-center flex-shrink-0 mx-3">
                                <div className="display-6 fw-bold">VS</div>
                                {renderStatus(g.game_status)}
                            </div>

                            {/* Команда B */}
                            <div
                                onClick={() => handleTeamClick(g.team_b_id)}
                                className={`team-card text-center flex-fill ms-2 clickable ${
                                    g.game_status === 2 && g.team_b_score > g.team_a_score
                                        ? "winner"
                                        : g.game_status === 2 && g.team_b_score < g.team_a_score
                                        ? "loser"
                                        : ""
                                }`}
                            >
                                <h5>{g.team_b_name}</h5>
                                <div className="score">{g.team_b_score}</div>
                            </div>
                        </Card.Body>
                    </Card>
                ))
            )}
        </Container>
    );
};

export default GamesPage;
