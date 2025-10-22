import { useEffect, useState } from "react";
import { Alert, Card, Container, Button, Modal, Form } from "react-bootstrap";
import { axiosDeleteEvent, axiosGetAllTournaments, axiosUpdateTournament } from "../../requests/tournamentRequests";
import { EventStatus, type EventResponse, type EventUpdate } from "../../types/TournamentTypes";
import formatDateForInput from "../../pkg/dateFormater";

const Tournaments: React.FC = () => {
    const [notify, setNotify] = useState<{ variant: "success" | "danger"; text: string } | null>(null);
    const [tournaments, setTournaments] = useState<EventResponse[]>([]);
    const [selectedTournament, setSelectedTournament] = useState<EventResponse | null>(null);
    const [showModal, setShowModal] = useState(false);

    const [formData, setFormData] = useState<EventUpdate>({
        name: "",
        start_date: "",
        end_date: "",
        status: 1,
    });

    useEffect(() => {
        const getTournaments = async () => {
            try {
                const res = await axiosGetAllTournaments();
                setTournaments(res);
            } catch (err) {
                const message = err instanceof Error ? err.message : "Не удалось загрузить турниры";
                setNotify({ variant: "danger", text: message });
            }
        };
        getTournaments();
    }, []);

    const handleEditClick = (tournament: EventResponse) => {
        console.log("TEST", tournament)
        setSelectedTournament(tournament);
        setFormData({
            name: tournament.name,
            start_date:formatDateForInput(tournament.start_date),
            end_date: formatDateForInput(tournament.end_date),
            status: tournament.status,
        });
        setShowModal(true);
    };

    const handleClose = () => {
        setShowModal(false);
        setSelectedTournament(null);
    };

    const handleChange = (e: React.ChangeEvent<
        HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement
    >) => {
        const { name, value } = e.target;
        setFormData((prev) => ({
            ...prev,
            [name]: name === "status" ? parseInt(value) : value,
        }));
    };

    const handleSave = async () => {
        if (!selectedTournament) return;

        const dto: EventUpdate = {
            name: formData.name,
            start_date: formData.start_date,
            end_date: formData.end_date,
            status: formData.status,
        };

        try {
            await axiosUpdateTournament(selectedTournament.id, dto);
            setNotify({ variant: "success", text: "Турнир успешно обновлён" });
            setShowModal(false);

            // обновляем список
            const updated = tournaments.map((t) =>
                t.id === selectedTournament.id ? { ...t, ...formData } : t
            );
            setTournaments(updated);
        } catch (err) {
            const message = err instanceof Error ? err.message : "Ошибка при обновлении турнира";
            setNotify({ variant: "danger", text: message });
        }
    };

    const handleDelete = async (id: string) => {
        try {
            await axiosDeleteEvent(id)
            setTournaments((prev) => prev.filter((t) => t.id !== id));
            setNotify({ variant: "success", text: "Турнир успешно удалён" })
        } catch (err) {
            const message = err instanceof Error ? err.message : "Ошибка при обновлении турнира";
            setNotify({ variant: "danger", text: message });
        }
    }

    return (
        <Container>
            <h2 className="mb-4 fw-bold text-start">Список всех турниров</h2>

            {notify && <Alert variant={notify.variant}>{notify.text}</Alert>}

            {tournaments.map((tournament) => (
                <Card key={tournament.id} className="rounded-3 mb-3 liqued-glass-bg">
                    <div className="d-flex justify-content-between">
                        <div className="ms-3 text-start">
                            <h5>{tournament.name}</h5>
                            <p>Статус: {tournament.status === EventStatus.Active ? "Активен" : "Завершён"}</p>
                        </div>
                        <div className="me-3 d-flex gap-3">
                            <button
                                className="btn btn-outline-primary"
                                onClick={() => handleEditClick(tournament)}
                            >
                                Редактировать
                            </button>
                            <button
                                className="btn btn-outline-danger"
                                onClick={() => handleDelete(tournament.id)}
                            >
                                Удалить
                            </button>
                        </div>
                    </div>
                </Card>
            ))}

            {/* Модальное окно редактирования */}
            <Modal show={showModal} onHide={handleClose} centered>
                <Modal.Header closeButton>
                    <Modal.Title>Редактировать турнир</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <Form>
                        <Form.Group className="mb-3">
                            <Form.Label>Название</Form.Label>
                            <Form.Control
                                type="text"
                                name="name"
                                value={formData.name}
                                onChange={handleChange}
                            />
                        </Form.Group>

                        <Form.Group className="mb-3">
                            <Form.Label>Дата начала</Form.Label>
                            <Form.Control
                                type="date"
                                name="start_date"
                                value={formData.start_date}
                                onChange={handleChange}
                            />
                        </Form.Group>

                        <Form.Group className="mb-3">
                            <Form.Label>Дата окончания</Form.Label>
                            <Form.Control
                                type="date"
                                name="end_date"
                                value={formData.end_date}
                                onChange={handleChange}
                            />
                        </Form.Group>

                        <Form.Group className="mb-3">
                            <Form.Label>Статус</Form.Label>
                            <Form.Select
                                name="status"
                                value={formData.status}
                                onChange={handleChange}
                            >
                                <option value={1}>Активен</option>
                                <option value={0}>Завершён</option>
                            </Form.Select>
                        </Form.Group>
                    </Form>
                </Modal.Body>
                <Modal.Footer>
                    <Button variant="secondary" onClick={handleClose}>
                        Отмена
                    </Button>
                    <Button variant="primary" onClick={handleSave}>
                        Сохранить
                    </Button>
                </Modal.Footer>
            </Modal>
        </Container>
    );
};

export default Tournaments;
