import React, { useState } from "react";
import { Container, Row, Col, Form, Button, Card, Table, Modal } from "react-bootstrap";

// Мок-данные школ
const schools: Record<string, string> = {
  "11111111-1111-1111-1111-111111111111": "Школа №1",
  "22222222-2222-2222-2222-222222222222": "Школа №2",
  "33333333-3333-3333-3333-333333333333": "Лицей №3",
};

interface Participant {
  fio: string;
  gender: number;
  class: number;
  school_id: string;
}

interface Team {
  team_name: string;
  responsible_name: string;
  responsible_number: string;
  school_id: string;
  event_id?: string; // будет браться из ссылки
}

export default function RegisterTeamPage() {
  const [team, setTeam] = useState<Team>({
    team_name: "",
    responsible_name: "",
    responsible_number: "",
    school_id: "",
    event_id: "", // можно потом заполнить из ссылки
  });

  const [participants, setParticipants] = useState<Participant[]>([]);
  const [showModal, setShowModal] = useState(false);
  const [currentParticipant, setCurrentParticipant] = useState<Omit<Participant, "school_id">>({
    fio: "",
    gender: 0,
    class: 0,
  });
  const [editIndex, setEditIndex] = useState<number | null>(null);

  const handleAddParticipant = () => {
    if (!team.school_id) {
      alert("Сначала выберите школу команды!");
      return;
    }

    const participant: Participant = {
      ...currentParticipant,
      school_id: team.school_id, // переносим выбранную школу команды
    };

    if (editIndex !== null) {
      const updated = [...participants];
      updated[editIndex] = participant;
      setParticipants(updated);
      setEditIndex(null);
    } else {
      setParticipants([...participants, participant]);
    }

    setCurrentParticipant({ fio: "", gender: 0, class: 0 });
    setShowModal(false);
  };

  const handleEditParticipant = (index: number) => {
    const { school_id, ...rest } = participants[index];
    setCurrentParticipant(rest);
    setEditIndex(index);
    setShowModal(true);
  };

  const handleDeleteParticipant = (index: number) => {
    setParticipants(participants.filter((_, i) => i !== index));
  };

  const handleSubmit = async () => {
    const dto = {
      ...team,
      participants,
    };

    console.log("Отправка:", dto);

    await fetch("/api/team/register", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(dto),
    });
  };

  return (
    <Container className="my-4">
      <h2 className="text-center mb-4">Регистрация команды</h2>

      {/* --- Данные команды --- */}
      <Card className="mb-4 shadow-sm">
        <Card.Body>
          <h5 className="mb-3">Информация о команде</h5>
          <Form>
            {/* Название команды на всю ширину */}
            <Row>
              <Col>
                <Form.Group className="mb-3">
                  <Form.Label>Название команды</Form.Label>
                  <Form.Control
                    type="text"
                    value={team.team_name}
                    onChange={(e) =>
                      setTeam({ ...team, team_name: e.target.value })
                    }
                  />
                </Form.Group>
              </Col>
            </Row>

            {/* Ответственный и телефон */}
            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Ответственный</Form.Label>
                  <Form.Control
                    type="text"
                    value={team.responsible_name}
                    onChange={(e) =>
                      setTeam({ ...team, responsible_name: e.target.value })
                    }
                  />
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Телефон ответственного</Form.Label>
                  <Form.Control
                    type="text"
                    value={team.responsible_number}
                    onChange={(e) =>
                      setTeam({ ...team, responsible_number: e.target.value })
                    }
                  />
                </Form.Group>
              </Col>
            </Row>

            {/* Выбор школы */}
            <Form.Group className="mb-3">
              <Form.Label>Школа</Form.Label>
              <Form.Select
                value={team.school_id}
                onChange={(e) =>
                  setTeam({ ...team, school_id: e.target.value })
                }
              >
                <option value="">Выберите школу</option>
                {Object.entries(schools).map(([id, name]) => (
                  <option key={id} value={id}>
                    {name}
                  </option>
                ))}
              </Form.Select>
            </Form.Group>
          </Form>
        </Card.Body>
      </Card>

      {/* --- Участники --- */}
      <Card className="mb-4 shadow-sm">
        <Card.Body>
          <div className="d-flex justify-content-between align-items-center mb-3">
            <h5>Участники</h5>
            <Button onClick={() => setShowModal(true)}>+ Добавить</Button>
          </div>

          <Table striped bordered hover>
            <thead>
              <tr>
                <th>ФИО</th>
                <th>Пол</th>
                <th>Класс</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {participants.map((p, i) => (
                <tr key={i}>
                  <td>{p.fio}</td>
                  <td>{p.gender === 1 ? "Мужской" : "Женский"}</td>
                  <td>{p.class}</td>
                  <td>
                    <Button
                      size="sm"
                      variant="outline-primary"
                      className="me-2"
                      onClick={() => handleEditParticipant(i)}
                    >
                      Редактировать
                    </Button>
                    <Button
                      size="sm"
                      variant="outline-danger"
                      onClick={() => handleDeleteParticipant(i)}
                    >
                      Удалить
                    </Button>
                  </td>
                </tr>
              ))}
              {participants.length === 0 && (
                <tr>
                  <td colSpan={5} className="text-center">
                    Участников пока нет
                  </td>
                </tr>
              )}
            </tbody>
          </Table>
        </Card.Body>
      </Card>

      {/* --- Кнопка отправки --- */}
      <div className="text-center">
        <Button size="lg" onClick={handleSubmit}>
          Отправить заявку
        </Button>
      </div>

      {/* --- Модалка участника --- */}
      <Modal show={showModal} onHide={() => setShowModal(false)} centered>
        <Modal.Header closeButton>
          <Modal.Title>
            {editIndex !== null ? "Редактировать участника" : "Добавить участника"}
          </Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <Form>
            <Form.Group className="mb-3">
              <Form.Label>ФИО</Form.Label>
              <Form.Control
                type="text"
                value={currentParticipant.fio}
                onChange={(e) =>
                  setCurrentParticipant({ ...currentParticipant, fio: e.target.value })
                }
              />
            </Form.Group>

            <Form.Group className="mb-3">
              <Form.Label>Пол</Form.Label>
              <Form.Select
                value={currentParticipant.gender}
                onChange={(e) =>
                  setCurrentParticipant({ ...currentParticipant, gender: parseInt(e.target.value) })
                }
              >
                <option value={0}>Женский</option>
                <option value={1}>Мужской</option>
              </Form.Select>
            </Form.Group>

            <Form.Group className="mb-3">
              <Form.Label>Класс</Form.Label>
              <Form.Control
                type="number"
                value={currentParticipant.class}
                onChange={(e) =>
                  setCurrentParticipant({ ...currentParticipant, class: parseInt(e.target.value) })
                }
              />
            </Form.Group>
          </Form>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => setShowModal(false)}>
            Отмена
          </Button>
          <Button onClick={handleAddParticipant}>
            {editIndex !== null ? "Сохранить" : "Добавить"}
          </Button>
        </Modal.Footer>
      </Modal>
    </Container>
  );
}
