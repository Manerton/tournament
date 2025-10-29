import React, { useEffect, useState, useCallback, useContext } from "react";
import { Container, Card, Badge, Spinner, Alert, Row, Col } from "react-bootstrap";
import { useParams, useNavigate } from "react-router-dom";
import { axiosGetGames } from "../../requests/gamesRequests";
import type { GameAllInformation } from "../../types/gameType";
import "../../assets/GamesPage.css";
import { useAuth } from "../helpers/AuthContext";
import { ThemeContext } from "../helpers/ThemeContext";

const REFRESH_INTERVAL = 2000; // обновляем каждые 2 секунды

const GamesPage: React.FC = () => {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();

    const [games, setGames] = useState<GameAllInformation[]>([]);
    const [loading, setLoading] = useState(true);
    const { initialized } = useAuth();

    const { theme } = useContext(ThemeContext);


    const loadGames = useCallback(async () => {
        try {
            const res = await axiosGetGames(id!);
            setGames((prev) => {
                const changed = JSON.stringify(prev) !== JSON.stringify(res);
                return changed ? res : prev;
            });
        } catch (err) {
            console.error("Ошибка загрузки игр:", err);
        } finally {
            setLoading(false);
        }
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
                return <Badge bg={theme === "light" ? "primary" : "warning"
                }> В игре</Badge >;
            default:
                return <Badge bg="success">Завершена</Badge>;
        }
    };

    const handleTeamClick = (game: GameAllInformation) => {
        if (!initialized) return;
        navigate(`/vote`, {state: {game}});
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
                <Row xs={1} md={2} className="g-4">
                    {games.map((g) => (
                        <Col key={g.team_a_id + g.team_b_id}>
                            <Card className="rounded-3 shadow-sm liquid-glass-bg" style={{ cursor: "pointer" }}>
                                <Card.Body className="d-flex justify-content-between align-items-center py-3 px-2 flex-wrap" 
                                   onClick={() => handleTeamClick(g)}>
                                    {/* Команда A */}
                                    <div
                                     
                                        className={`team-card text-center flex-fill me-2 clickable ${g.game_status === 2 && g.team_a_score > g.team_b_score
                                            ? "winner"
                                            : g.game_status === 2 && g.team_a_score < g.team_b_score
                                                ? "loser"
                                                : ""
                                            }`}
                                    >
                                        <h5>{g.team_a_name}</h5>
                                        <div className="score">{g.team_a_score}</div>
                                    </div>

                                    {/* VS и статус */}
                                    <div className="vs-container text-center flex-shrink-0 mx-3">
                                        <div className="display-6 fw-bold text-danger">VS</div>
                                        {renderStatus(g.game_status)}
                                    </div>

                                    {/* Команда B */}
                                    <div
                                        className={`team-card text-center flex-fill ms-2 border-1 clickable ${g.game_status === 2 && g.team_b_score > g.team_a_score
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
                        </Col>
                    ))}
                </Row>
            )}
        </Container>
    );
};

export default GamesPage;
