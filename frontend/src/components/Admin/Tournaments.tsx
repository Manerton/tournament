import React, { useEffect, useState } from "react";
import { Alert, Card, Container, Button, Modal, Form } from "react-bootstrap";
import {
    axiosDeleteEvent,
    axiosGetAllTournaments,
    axiosUpdateTournament,
} from "../../requests/tournamentRequests";
import { axiosCreateTournament } from "../../requests/tournamentRequests";
import {
    EventStatus,
    type EventResponse,
    type EventUpdate,
} from "../../types/TournamentTypes";
import formatDateForInput from "../../pkg/dateFormater";

const Tournaments: React.FC = () => {
    const [notify, setNotify] = useState<{ variant: "success" | "danger"; text: string } | null>(
        null
    );
    const [tournaments, setTournaments] = useState<EventResponse[]>([]);
    const [selectedTournament, setSelectedTournament] = useState<EventResponse | null>(null);
    const [showModal, setShowModal] = useState(false);
    const [isCreating, setIsCreating] = useState(false);

    const [formData, setFormData] = useState({
        name: "",
        start_date: "",
        end_date: "",
        status: 1,
        type_draws: 0,
    });

    const getTournaments = async () => {
        try {
            const res = await axiosGetAllTournaments();
            setTournaments(res);
        } catch (err) {
            const message = err instanceof Error ? err.message : "Не удалось загрузить турниры";
            setNotify({ variant: "danger", text: message });
        }
    };

    useEffect(() => {
        getTournaments();
    }, []);

    const handleEditClick = (tournament: EventResponse) => {
        setSelectedTournament(tournament);
        setIsCreating(false);
        setFormData({
            name: tournament.name,
            start_date: formatDateForInput(tournament.start_date),
            end_date: formatDateForInput(tournament.end_date),
            status: tournament.status,
            type_draws: tournament.type_draws ?? 0,
        });
        setShowModal(true);
    };

    const handleCreateClick = () => {
        setSelectedTournament(null);
        setIsCreating(true);
        setFormData({
            name: "",
            start_date: "",
            end_date: "",
            status: EventStatus.Registration,
            type_draws: 0,
        });
        setShowModal(true);
    };

    const handleClose = () => {
        setShowModal(false);
        setSelectedTournament(null);
    };

    const handleChange = (
        e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>
    ) => {
        const { name, value } = e.target;
        setFormData((prev) => ({
            ...prev,
            [name]: name === "status" || name === "type_draws" ? parseInt(value) : value,
        }));
    };

    const handleSave = async () => {
        try {
            if (isCreating) {
                const dto = {
                    name: formData.name,
                    start_date: formData.start_date,
                    end_date: formData.end_date,
                    type_draws: formData.type_draws,
                };
                await axiosCreateTournament(dto);
                getTournaments();
                setNotify({ variant: "success", text: "Турнир успешно создан" });
            } else if (selectedTournament) {
                const dto: EventUpdate = {
                    name: formData.name,
                    start_date: formData.start_date,
                    end_date: formData.end_date,
                    status: formData.status,
                };
                await axiosUpdateTournament(selectedTournament.id, dto);
                getTournaments();

                setNotify({ variant: "success", text: "Турнир успешно обновлён" });
            }

            setShowModal(false);
        } catch (err) {
            const message =
                err instanceof Error
                    ? err.message
                    : isCreating
                        ? "Ошибка при создании турнира"
                        : "Ошибка при обновлении турнира";
            setNotify({ variant: "danger", text: message });
        }
    };

    const handleDelete = async (id: string) => {
        try {
            await axiosDeleteEvent(id);
            setTournaments((prev) => prev.filter((t) => t.id !== id));
            setNotify({ variant: "success", text: "Турнир успешно удалён" });
        } catch (err) {
            const message = err instanceof Error ? err.message : "Ошибка при удалении турнира";
            setNotify({ variant: "danger", text: message });
        }
    };

    return (
        <Container>
            <div className="d-flex justify-content-between align-items-center mb-4">
                <h2 className="fw-bold">Список всех турниров</h2>
                <Button variant="success" onClick={handleCreateClick}>
                    + Создать турнир
                </Button>
            </div>

            {notify && <Alert variant={notify.variant}>{notify.text}</Alert>}

            {tournaments.map((tournament) => (
                <Card key={tournament.id} className="rounded-3 mb-3 liquid-glass-bg">
                    <div className="d-flex justify-content-between">
                        <div className="ms-3 text-start">
                            <h5>{tournament.name}</h5>
                            <p>
                                Даты: c {formatDateForInput(tournament.start_date)} по{" "}
                                {formatDateForInput(tournament.end_date)}
                            </p>
                            Статус:{" "}
                            <span
                                className={`badge ${tournament.status === EventStatus.Active
                                    ? "bg-success"
                                    : tournament.status === EventStatus.Registration
                                        ? "bg-warning text-dark"
                                        : "bg-secondary"
                                    }`}
                            >
                                {tournament.status === EventStatus.Active
                                    ? "Активен"
                                    : tournament.status === EventStatus.Registration
                                        ? "Регистрация"
                                        : "Завершён"}
                            </span>
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

            {/* Модальное окно создания/редактирования */}
            <Modal show={showModal} onHide={handleClose} centered>
                <Modal.Header closeButton>
                    <Modal.Title>
                        {isCreating ? "Создать турнир" : "Редактировать турнир"}
                    </Modal.Title>
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

                        {isCreating && (
                            <Form.Group className="mb-3">
                                <Form.Label>Тип жеребьёвки</Form.Label>
                                <Form.Select
                                    name="type_draws"
                                    value={formData.type_draws}
                                    onChange={handleChange}
                                >
                                    <option value={0}>Не указано</option>
                                    <option value={3}>Швейцарская</option>
                                </Form.Select>
                            </Form.Group>
                        )}

                        {!isCreating && (
                            <Form.Group className="mb-3">
                                <Form.Label>Статус</Form.Label>
                                <Form.Select
                                    name="status"
                                    value={formData.status}
                                    onChange={handleChange}
                                >
                                    <option value={EventStatus.Active}>Активен</option>
                                    <option value={EventStatus.Closed}>Завершён</option>
                                    <option value={EventStatus.Registration}>Регистрация</option>
                                </Form.Select>
                            </Form.Group>
                        )}
                    </Form>
                </Modal.Body>
                <Modal.Footer>
                    <Button variant="secondary" onClick={handleClose}>
                        Отмена
                    </Button>
                    <Button variant="primary" onClick={handleSave}>
                        {isCreating ? "Создать" : "Сохранить"}
                    </Button>
                </Modal.Footer>
            </Modal>
        </Container>
    );
};

export default Tournaments;
