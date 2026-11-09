import { useEffect, useState } from "react";
import { Button, Table } from "react-bootstrap";
import { useNavigate, useParams } from "react-router-dom";
import { axiosQuizGetQuestions } from "../../requests/quizRequests";
import type { QuizQuestion } from "../../types/quizTypes";
import "../../assets/Quiz.css";

export default function QuizQuestionBank() {
    const { eventId } = useParams<{ eventId: string }>();
    const navigate = useNavigate();
    const [questions, setQuestions] = useState<QuizQuestion[]>([]);
    const [error, setError] = useState("");

    useEffect(() => {
        if (!eventId) return;
        axiosQuizGetQuestions(eventId)
            .then(setQuestions)
            .catch((e) => setError(e.message));
    }, [eventId]);

    if (!eventId) return null;

    const rounds = [
        { round: 1, name: "Раунд 1 — «Стрелы логики»" },
        { round: 2, name: "Раунд 2 — «Ударный вектор»" },
        { round: 3, name: "Раунд 3 — «Формула клинка»" },
    ];

    return (
        <div className="quiz-admin-root">
            <div className="d-flex justify-content-between align-items-center mb-4">
                <h2 className="mb-0">Банк вопросов</h2>
                <Button variant="outline-primary" onClick={() => navigate(`/quiz/admin/${eventId}`)}>
                    ← Назад к панели ведущего
                </Button>
            </div>

            {error && <div className="quiz-msg error mb-3">{error}</div>}

            {rounds.map(({ round, name }) => {
                const list = questions.filter((q) => q.round === round);
                return (
                    <div className="quiz-admin-section liquid-glass-bg" key={round}>
                        <h4>{name}</h4>
                        {list.length === 0 ? (
                            <span className="text-body-secondary">Вопросы не загружены.</span>
                        ) : (
                            <Table striped bordered hover size="sm">
                                <thead>
                                    <tr>
                                        <th style={{ width: "60px" }}>№</th>
                                        <th>Вопрос</th>
                                        <th>Правильный ответ</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {list.map((q) => (
                                        <tr key={q.id}>
                                            <td>{q.question_number}</td>
                                            <td>{q.text}</td>
                                            <td>{q.answer}</td>
                                        </tr>
                                    ))}
                                </tbody>
                            </Table>
                        )}
                    </div>
                );
            })}
        </div>
    );
}
