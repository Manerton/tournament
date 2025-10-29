import { Alert, Accordion, Button, Container, Spinner, Modal } from "react-bootstrap";
import type { GameAllInformation } from "../../types/gameType";
import { useEffect, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { GetTeamMembers } from "../../requests/teamRequests";
import type { ParticipantResponseDTO } from "../../types/participantTypes";
import { axiosUpdateParticipantScore } from "../../requests/participantRequest";

interface Props {
    game: GameAllInformation;
}

const GameVotePage: React.FC = () => {
    const location = useLocation();
    const navigate = useNavigate();

    const [finishing, setFinishing] = useState(false);
    const [showConfirm, setShowConfirm] = useState(false);


    const { game } = (location.state || {}) as Props;
    const [loading, setLoading] = useState(true);
    const [teamA, setTeamA] = useState<ParticipantResponseDTO[]>([]);
    const [teamB, setTeamB] = useState<ParticipantResponseDTO[]>([]);

    useEffect(() => {
        if (!game) return;

        const loadTeams = async () => {
            try {
                const resTeamA = await GetTeamMembers(game.team_a_id);
                const resTeamB = await GetTeamMembers(game.team_b_id);
                setTeamA(resTeamA);
                setTeamB(resTeamB);
            } catch (err) {
                console.error("Ошибка загрузки команд:", err);
            } finally {
                setLoading(false);
            }
        };
        loadTeams();
    }, [game]);

    const handleFinishGame = async () => {
        setFinishing(true);
        try {
            // await axiosFinishGame(game.id); // TODO: реализовать позже
            console.log(`Игра ${game.id} завершена`);
            navigate(-1);
        } catch (err) {
            console.error("Ошибка при завершении игры:", err);
        } finally {
            setFinishing(false);
            setShowConfirm(false);
        }
    };

    const handleScoreChange = async (participantId: string, delta: number) => {
        try {
            await axiosUpdateParticipantScore(game.id, participantId, delta);
        } catch (err) {
            console.error("Ошибка обновления очков:", err);
        }
    };

    if (!game) {
        return (
            <Container className="py-4">
                <Alert variant="danger" className="mt-4 text-center">
                    Нет данных об игре. Вернуться{" "}
                    <Alert.Link onClick={() => navigate(-1)}>назад</Alert.Link>
                </Alert>
            </Container>
        );
    }

    const renderTeam = (teamName: string, team: ParticipantResponseDTO[], key: string) => (
        <Accordion.Item eventKey={key} className="liquid-glass-bg">
            <Accordion.Header >{teamName}</Accordion.Header>
            <Accordion.Body>
                {team.length === 0 ? (
                    <Alert variant="light">Нет участников</Alert>
                ) : (
                    team.map((p) => (
                        <div
                            key={p.id}
                            className="d-flex justify-content-between align-items-center border-bottom border-black py-2 flex-wrap"
                        >
                            <div>
                                <strong>{p.FIO}</strong>
                            </div>
                            <div className="d-flex flex-column align-items-end gap-1">
                                <div className="d-flex gap-2">
                                    {[1, 3, 5].map((val) => (
                                        <Button
                                            key={`plus-${val}-${p.id}`}
                                            size="sm"
                                            variant="outline-success"
                                            onClick={() => handleScoreChange(p.id, val)}
                                        >
                                            +{val}
                                        </Button>
                                    ))}
                                </div>
                                <div className="d-flex gap-2">
                                    {[1, 3, 5].map((val) => (
                                        <Button
                                            key={`minus-${val}-${p.id}`}
                                            size="sm"
                                            variant="outline-danger"
                                            onClick={() => handleScoreChange(p.id, -val)}
                                        >
                                            -{val}
                                        </Button>
                                    ))}
                                </div>
                            </div>
                        </div>
                    ))
                )}
            </Accordion.Body>
        </Accordion.Item>
    );

    return (
        <Container className="py-4">
            <h2 className="mb-4 fw-bold text-center display-6">Оценка игр турнира</h2>

            <div className="d-flex justify-content-center flex-wrap align-items-center mb-4">
                <h4 className="fw-bold me-3">{game.team_a_name}</h4>
                <h4 className="text-danger">VS</h4>
                <h4 className="fw-bold ms-3">{game.team_b_name}</h4>
            </div>

            {loading ? (
                <div className="text-center my-5">
                    <Spinner animation="border" />
                </div>
            ) : (
                <>
                    <Accordion defaultActiveKey="0" className="liquid-glass-bg">
                        {renderTeam(game.team_a_name, teamA, "0")}
                        {renderTeam(game.team_b_name, teamB, "1")}
                    </Accordion>

                    <div className="text-center mt-5">
                        <Button
                            variant="outline-danger"
                            size="lg"
                            onClick={() => setShowConfirm(true)}
                        >
                            Завершить игру
                        </Button>
                    </div>
                </>
            )}

            {/* Модальное окно подтверждения */}
            <Modal show={showConfirm} onHide={() => setShowConfirm(false)} centered>
                <Modal.Header closeButton>
                    <Modal.Title>Подтверждение действия</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    Вы уверены, что хотите <strong>завершить игру</strong>?
                    После этого редактирование очков будет недоступно.
                </Modal.Body>
                <Modal.Footer>
                    <Button variant="secondary" onClick={() => setShowConfirm(false)}>
                        Отмена
                    </Button>
                    <Button
                        variant="danger"
                        onClick={handleFinishGame}
                        disabled={finishing}
                    >
                        {finishing ? "Завершаем..." : "Да, завершить"}
                    </Button>
                </Modal.Footer>
            </Modal>
        </Container>
    );
};

export default GameVotePage;
