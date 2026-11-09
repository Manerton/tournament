import { useState } from "react";
import { Alert, Button, Card, Container, Spinner, Table } from "react-bootstrap";
import { useNavigate } from "react-router-dom";
import { useQuizState } from "../helpers/useQuizState";
import { useAuth } from "../helpers/AuthContext";
import { UserRole } from "../../types/userType";
import { axiosStartTournament } from "../../requests/tournamentRequests";
import { Round3OrderNames } from "../../types/quizTypes";
import "../../assets/Quiz.css";

export default function QuizToursPanel({ eventId }: { eventId: string }) {
    const navigate = useNavigate();
    const { state } = useQuizState(eventId);
    const { user } = useAuth();
    const [starting, setStarting] = useState(false);
    const [error, setError] = useState("");

    const handleStart = async () => {
        setStarting(true);
        setError("");
        try {
            // For quiz events this provisions quiz teams (access codes) for accepted teams.
            await axiosStartTournament(eventId);
        } catch (e: any) {
            setError(e.message || "Не удалось начать турнир");
        } finally {
            setStarting(false);
        }
    };

    const teams = state
        ? [...state.teams].sort((a, b) => b.total_score - a.total_score || a.number - b.number)
        : [];

    return (
        <Container className="py-4">
            <h2 className="mb-4 fw-bold text-center display-6">Квиз-шоу</h2>

            {error && <Alert variant="danger">{error}</Alert>}

            <div className="d-flex justify-content-around gap-3 mb-4 flex-wrap">
                {user?.role === UserRole.Admin && (
                    <Button
                        variant="primary"
                        size="lg"
                        className="flex-grow-1"
                        onClick={() => navigate(`/quiz/admin/${eventId}`)}
                    >
                        Панель ведущего
                    </Button>
                )}
                <Button
                    variant="outline-primary"
                    size="lg"
                    className="flex-grow-1"
                    onClick={() => navigate(`/quiz/display/${eventId}`)}
                >
                    Экран для зрителей
                </Button>
                <Button
                    variant="outline-primary"
                    size="lg"
                    className="flex-grow-1"
                    onClick={() => navigate(`/quiz/herald/${eventId}`)}
                >
                    Экран герольда
                </Button>
            </div>

            <Card className="rounded-4 shadow-sm liquid-glass-bg mb-4">
                <Card.Body className="py-4">
                    <div className="d-flex justify-content-between align-items-center mb-3">
                        <Card.Title className="fw-bold fs-3 text-body mb-0">
                            Живые результаты
                        </Card.Title>
                        <Button
                            variant="outline-success"
                            disabled={starting}
                            onClick={handleStart}
                        >
                            {starting ? (
                                <>
                                    <Spinner as="span" animation="border" size="sm" className="me-2" />
                                    Запуск...
                                </>
                            ) : (
                                "Начать турнир / обновить команды"
                            )}
                        </Button>
                    </div>

                    {!state ? (
                        <Spinner animation="border" />
                    ) : teams.length === 0 ? (
                        <Alert variant="info" className="mb-0">
                            Пока нет команд. Примите заявки на странице «Заявки команд», затем
                            нажмите «Начать турнир / обновить команды» — для каждой команды будет
                            сгенерирован код доступа для герольда.
                        </Alert>
                    ) : (
                        <>
                            <div className="mb-2 text-body-secondary">
                                {state.round_active
                                    ? `Идёт раунд ${state.current_round}` +
                                      (state.current_question > 0
                                          ? `, вопрос ${state.current_question}`
                                          : "")
                                    : "Раунд не начат"}
                            </div>
                            <Table striped bordered hover size="sm" className="mb-0">
                                <thead>
                                    <tr>
                                        <th>№</th>
                                        <th>Команда</th>
                                        <th>Раунд 1</th>
                                        <th>Раунд 2</th>
                                        <th>Раунд 3</th>
                                        <th>Всего</th>
                                        <th>Орден</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {teams.map((t) => (
                                        <tr key={t.team_id}>
                                            <td>{t.number}</td>
                                            <td>{t.name}</td>
                                            <td>{t.round1_score}</td>
                                            <td>{t.round2_score}</td>
                                            <td>{t.round3_score}</td>
                                            <td>
                                                <strong>{t.total_score}</strong>
                                            </td>
                                            <td>
                                                {t.round3_order > 0
                                                    ? Round3OrderNames[t.round3_order] +
                                                      (t.round3_eliminated ? " (выбыла)" : "")
                                                    : "—"}
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </Table>
                        </>
                    )}
                </Card.Body>
            </Card>
        </Container>
    );
}
