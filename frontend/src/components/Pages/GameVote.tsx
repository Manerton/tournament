import { Alert, Accordion, Button, Container, Spinner, Modal } from "react-bootstrap";
import { GameStatus, type GameAllInformation } from "../../types/gameType";
import { useEffect, useRef, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { GetTeamMembers } from "../../requests/teamRequests";
import type { ParticipantResponseDTO } from "../../types/participantTypes";
import { axiosUpdateParticipantScore } from "../../requests/participantRequest";
import { axiosFinishGame, axiosGetGameParticipantsByListId } from "../../requests/gamesRequests";
import { axiosGetGameById } from "../../requests/gamesRequests";
import SubmissionSidebar from "../common/SubmissionSidebar";

interface Props {
  id: string;
  mainEventId: string;
}

const GameVotePage: React.FC = () => {
  const location = useLocation();
  const navigate = useNavigate();

  const [finishing, setFinishing] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);
  const [notify, setNotify] = useState<{ variant: "success" | "danger"; text: string } | null>(null);
  const alertRef = useRef<HTMLDivElement | null>(null);

  const { id } = (location.state || {}) as Props;
  const { mainEventId } = (location.state || {}) as Props;

  const [loading, setLoading] = useState(true);
  const [teamA, setTeamA] = useState<ParticipantResponseDTO[]>([]);
  const [teamB, setTeamB] = useState<ParticipantResponseDTO[]>([]);
  const [scores, setScores] = useState<Record<string, number>>({});
  const [game, setGame] = useState<GameAllInformation>();

  // ✅ Загрузка данных об игре
  const loadGame = async () => {
    try {
      const res = await axiosGetGameById(id);
      setGame(res);
    } catch (err) {
      console.error("Ошибка загрузки игры:", err);
      setNotify({ variant: "danger", text: "Не удалось загрузить данные об игре" });
    }
  };

  // ✅ Первоначальная загрузка и обновление каждые 1 секунду
  useEffect(() => {
    if (!id) return;

    loadGame(); // первая загрузка
    const interval = setInterval(loadGame, 1000); // обновление каждую секунду

    return () => clearInterval(interval); // очистка при размонтировании
  }, [id]);

  // уведомление
  useEffect(() => {
    if (notify && alertRef.current) {
      alertRef.current.scrollIntoView({ behavior: "smooth", block: "center" });
    }
  }, [notify]);

  // загрузка команд
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

  // загрузка индивидуальных очков участников
  useEffect(() => {
    const loadScores = async () => {
      try {
        const allparticipantIds = [...teamA, ...teamB].map((p) => p.id);
        const res = await axiosGetGameParticipantsByListId(allparticipantIds);
        const scoresMap: Record<string, number> = {};
        res.forEach((s) => (scoresMap[s.participant_id] = s.score));
        setScores(scoresMap);
      } catch (err) {
        const message =
          err instanceof Error ? err.message : "Не удалось загрузить очки участников";
        setNotify({ variant: "danger", text: message });
        console.error("Ошибка загрузки очков участников:", err);
      }
    };

    if (teamA.length || teamB.length) {
      loadScores();
    }
  }, [teamA, teamB]);

  const handleFinishGame = async () => {
    setFinishing(true);
    try {
      await axiosFinishGame(game!.id);
      console.log(`Игра ${game!.id} завершена`);
      navigate(-1);
    } catch (err) {
      const message = err instanceof Error ? err.message : "Не удалось завершить игру";
      setNotify({ variant: "danger", text: message });
      console.error("Ошибка при завершении игры:", err);
    } finally {
      setFinishing(false);
      setShowConfirm(false);
    }
  };

  const handleScoreChange = async (participantId: string, delta: number) => {
    try {
      const updateScore = await axiosUpdateParticipantScore(game!.id, participantId, delta);
      setScores((prev) => ({
        ...prev,
        [participantId]: updateScore,
      }));
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
      <Accordion.Header>{teamName}</Accordion.Header>
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
              <div className="text-muted small">Текущий счёт: {scores[p.id] ?? 0}</div>
              <div className="d-flex flex-column align-items-end gap-1">
                <div className="d-flex gap-2">
                  {[1, 3, 5].map((val) => (
                    <Button
                      key={`plus-${val}-${p.id}`}
                      size="sm"
                      variant="outline-success"
                      onClick={() => handleScoreChange(p.id, val)}
                      disabled={game.game_status !== GameStatus.NotFinished}
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
                      disabled={game.game_status !== GameStatus.NotFinished}
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
      <SubmissionSidebar eventId={mainEventId} />
      <h2 className="mb-4 fw-bold text-center display-6">Оценка игр тура</h2>

      {notify && (
        <div ref={alertRef}>
          <Alert variant={notify.variant} className="text-center">
            {notify.text}
          </Alert>
        </div>
      )}

      {/* ✅ блок со счётом игры */}
      <div className="text-center mb-4">
        <h4 className="fw-bold">
          {game.team_a_name} <span className="text-danger">VS</span> {game.team_b_name}
        </h4>
        <div className="fs-3 fw-bold mt-2">
          <span className="text-primary">{game.team_a_score}</span>
          <span className="mx-3 text-dark">:</span>
          <span className="text-success">{game.team_b_score}</span>
        </div>
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
            {game.game_status == GameStatus.NotFinished ? (
              <Button variant="outline-danger" size="lg" onClick={() => setShowConfirm(true)}>
                Завершить игру
              </Button>
            ) : (
              <div>Игра завершена</div>
            )}
          </div>
        </>
      )}

      {/* Модальное окно подтверждения */}
      <Modal show={showConfirm} onHide={() => setShowConfirm(false)} centered>
        <Modal.Header closeButton>
          <Modal.Title>Подтверждение действия</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          Вы уверены, что хотите <strong>завершить игру</strong>? После этого редактирование очков
          будет недоступно.
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => setShowConfirm(false)}>
            Отмена
          </Button>
          <Button variant="danger" onClick={handleFinishGame} disabled={finishing}>
            {finishing ? "Завершаем..." : "Да, завершить"}
          </Button>
        </Modal.Footer>
      </Modal>
    </Container>
  );
};

export default GameVotePage;
