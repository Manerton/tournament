import { useCallback, useEffect, useRef, useState } from "react";
import { Button, Table } from "react-bootstrap";
import { useNavigate, useParams } from "react-router-dom";
import { useQuizState } from "../helpers/useQuizState";
import {
    axiosQuizCloseQuestion,
    axiosQuizGetQuestions,
    axiosQuizGetTeams,
    axiosQuizGrantTurn,
    axiosQuizPauseTimer,
    axiosQuizResetQuestion,
    axiosQuizResumeTimer,
    axiosQuizRevealAnswer,
    axiosQuizSetupTeams,
    axiosQuizShowLeaderboard,
    axiosQuizStartQuestion,
    axiosQuizStartRound,
} from "../../requests/quizRequests";
import { BuzzStatus, Round3OrderNames, type QuizQuestion, type QuizTeamAdmin } from "../../types/quizTypes";
import "../../assets/Quiz.css";

function formatTime(totalSeconds: number): string {
    const m = Math.floor(totalSeconds / 60);
    const s = totalSeconds % 60;
    return `${m}:${s.toString().padStart(2, "0")}`;
}

export default function QuizAdmin() {
    const { eventId } = useParams<{ eventId: string }>();
    const navigate = useNavigate();
    const { state } = useQuizState(eventId);

    const [teams, setTeams] = useState<QuizTeamAdmin[]>([]);
    const [questions, setQuestions] = useState<QuizQuestion[]>([]);
    const [error, setError] = useState("");
    const [notice, setNotice] = useState("");

    const [remainingMs, setRemainingMs] = useState(0);
    const deadlineRef = useRef(0);

    const loadTeams = useCallback(() => {
        if (!eventId) return;
        axiosQuizGetTeams(eventId).then(setTeams).catch(() => setTeams([]));
    }, [eventId]);

    const loadQuestions = useCallback(() => {
        if (!eventId) return;
        axiosQuizGetQuestions(eventId).then(setQuestions).catch(() => setQuestions([]));
    }, [eventId]);

    useEffect(() => {
        loadTeams();
        loadQuestions();
    }, [loadTeams, loadQuestions]);

    useEffect(() => {
        if (!state) return;
        deadlineRef.current = Date.now() + state.timer.remaining_ms;
        setRemainingMs(state.timer.remaining_ms);
    }, [state]);

    useEffect(() => {
        const tick = setInterval(() => {
            setRemainingMs(Math.max(deadlineRef.current - Date.now(), 0));
        }, 250);
        return () => clearInterval(tick);
    }, []);

    const run = async (fn: () => Promise<void>, successMsg: string) => {
        setError("");
        setNotice("");
        try {
            await fn();
            setNotice(successMsg);
            loadTeams();
        } catch (e: any) {
            setError(e.message);
        }
    };

    if (!eventId) return null;

    const roundCfg = state?.rounds.find((r) => r.round === state.current_round);
    const nextQuestionNumber = (state?.current_question ?? 0) + 1;
    const nextQuestionExists = roundCfg ? nextQuestionNumber <= roundCfg.question_count : false;
    const currentQuestionData =
        state && roundCfg
            ? questions.find(
                  (q) => q.round === state.current_round && q.question_number === state.current_question,
              )
            : undefined;

    return (
        <div className="quiz-admin-root">
            <h2 className="mb-4">Панель ведущего — квиз-шоу</h2>

            {error && <div className="quiz-msg error mb-3">{error}</div>}
            {notice && <div className="quiz-msg ok mb-3">{notice}</div>}

            {/* Teams & access codes */}
            <div className="quiz-admin-section liquid-glass-bg">
                <h4>Команды и коды доступа</h4>
                <div className="quiz-admin-row mb-3">
                    <Button onClick={() => run(() => axiosQuizSetupTeams(eventId), "Команды подготовлены")}>
                        Подготовить / добавить команды
                    </Button>
                    <span className="text-body-secondary">
                        Коды выдаются герольдам для входа на экраны оценки.
                    </span>
                </div>
                {teams.length > 0 && (
                    <Table striped bordered hover size="sm">
                        <thead>
                            <tr>
                                <th>Команда</th>
                                <th>Код</th>
                                <th>Раунд 1</th>
                                <th>Раунд 2</th>
                                <th>Раунд 3</th>
                                <th>Орден</th>
                                <th></th>
                            </tr>
                        </thead>
                        <tbody>
                            {teams.map((t) => (
                                <tr key={t.id}>
                                    <td>{t.name}</td>
                                    <td>
                                        <strong>{t.access_code}</strong>
                                    </td>
                                    <td>{t.round1_score}</td>
                                    <td>{t.round2_score}</td>
                                    <td>{t.round3_score}</td>
                                    <td>
                                        {t.round3_order > 0
                                            ? Round3OrderNames[t.round3_order]
                                            : "—"}
                                        {t.round3_eliminated ? " (выбыла)" : ""}
                                    </td>
                                    <td>
                                        <Button
                                            size="sm"
                                            variant="warning"
                                            onClick={() =>
                                                run(
                                                    () => axiosQuizGrantTurn(eventId, t.team_id),
                                                    `Право ответа передано: ${t.name}`,
                                                )
                                            }
                                        >
                                            Дать право ответа
                                        </Button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </Table>
                )}
            </div>

            {/* Question bank (read-only: questions are preloaded via migrations) */}
            <div className="quiz-admin-section liquid-glass-bg">
                <div className="quiz-admin-row">
                    <h4 className="mb-0">Банк вопросов</h4>
                    <Button
                        variant="outline-secondary"
                        onClick={() => navigate(`/quiz/admin/${eventId}/questions`)}
                    >
                        Открыть банк вопросов
                    </Button>
                    <span className="text-body-secondary">
                        Вопросы загружаются в базу заранее (миграцией). Загружено: {questions.length}.
                    </span>
                </div>
            </div>

            {/* Game flow */}
            <div className="quiz-admin-section liquid-glass-bg">
                <h4>Управление игрой</h4>
                <div className="quiz-admin-row mb-3">
                    {[1, 2, 3].map((r) => (
                        <Button
                            key={r}
                            variant={state?.current_round === r ? "success" : "outline-secondary"}
                            onClick={() =>
                                run(() => axiosQuizStartRound(eventId, r), `Раунд ${r} начат`)
                            }
                        >
                            Начать раунд {r}
                        </Button>
                    ))}
                    <Button
                        variant={state?.leaderboard_shown ? "success" : "outline-secondary"}
                        onClick={() =>
                            run(
                                () => axiosQuizShowLeaderboard(eventId, !state?.leaderboard_shown),
                                "Таблица лидеров переключена",
                            )
                        }
                    >
                        {state?.leaderboard_shown ? "Скрыть таблицу" : "Показать таблицу"}
                    </Button>
                </div>

                <div className="quiz-admin-row mb-2">
                    <span className="quiz-timer-big">{formatTime(Math.ceil(remainingMs / 1000))}</span>
                    <span className="text-body-secondary">
                        {roundCfg
                            ? `Раунд ${roundCfg.round} — «${roundCfg.name}», вопрос ${state?.current_question ?? 0} из ${roundCfg.question_count}`
                            : "Раунд не начат"}
                        {state?.answer_revealed ? " · ответ показан" : ""}
                    </span>
                </div>

                <div className="quiz-admin-row mb-3">
                    <Button
                        variant="primary"
                        disabled={!roundCfg || !nextQuestionExists}
                        onClick={() => run(() => axiosQuizStartQuestion(eventId), `Вопрос ${nextQuestionNumber} запущен`)}
                    >
                        {`Запустить вопрос ${nextQuestionNumber}`}
                    </Button>
                    <Button
                        variant="outline-secondary"
                        onClick={() => run(() => axiosQuizPauseTimer(eventId), "Таймер приостановлен")}
                    >
                        Пауза
                    </Button>
                    <Button
                        variant="outline-secondary"
                        onClick={() => run(() => axiosQuizResumeTimer(eventId), "Таймер продолжен")}
                    >
                        Продолжить
                    </Button>
                    <Button
                        variant="outline-secondary"
                        onClick={() => run(() => axiosQuizCloseQuestion(eventId), "Вопрос закрыт")}
                    >
                        Закрыть вопрос
                    </Button>
                    <Button
                        variant="success"
                        onClick={() => run(() => axiosQuizRevealAnswer(eventId), "Ответ показан")}
                    >
                        Показать ответ
                    </Button>
                    <Button
                        variant="warning"
                        onClick={() => run(() => axiosQuizResetQuestion(eventId), "Очередь сброшена")}
                    >
                        Сбросить очередь
                    </Button>
                </div>

                {state?.current_round === 1 && state.current_question > 0 && (
                    <div className="mb-2">
                        <div className="text-body-secondary mb-2">
                            Очередь нажатий (попыток использовано: {state.attempts_used} /{" "}
                            {state.attempts_max})
                        </div>
                        {state.buzz_queue.length === 0 && (
                            <span className="text-body-secondary">Нажатий пока не было.</span>
                        )}
                        <div className="quiz-admin-row">
                            {state.buzz_queue.map((b, i) => (
                                <div
                                    key={b.team_id}
                                    className={`quiz-queue-chip${
                                        b.is_active ? " active" : ""
                                    }${b.status === BuzzStatus.Wrong ? " wrong" : ""}${
                                        b.status === BuzzStatus.Correct ? " correct" : ""
                                    }`}
                                >
                                    {i + 1}. {b.team_name}
                                </div>
                            ))}
                        </div>
                    </div>
                )}

                {state && roundCfg && state.current_question > 0 && (
                    <div className="mb-3" style={{ textAlign: "left" }}>
                        <div className="text-body-secondary mb-2">
                            Текущий вопрос для ведущего:
                        </div>
                        {currentQuestionData ? (
                            <div className="liquid-glass-bg rounded-3 p-3">
                                <div>{currentQuestionData.text}</div>
                                <div className="text-success mt-2">
                                    Ответ: {currentQuestionData.answer}
                                </div>
                            </div>
                        ) : (
                            <div className="quiz-msg error">
                                Для текущего вопроса нет данных в банке — проверьте загрузку вопросов.
                            </div>
                        )}
                    </div>
                )}
            </div>
        </div>
    );
}
