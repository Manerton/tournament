import React, { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { Container, Row, Col, Form, Button, Card, Table, Modal, Spinner, Alert } from "react-bootstrap";
import Select from "react-select";
import { API_URL } from "../../config/routes";

// ------------------------------------
// Типы DTO (согласно новому ТЗ)
// ------------------------------------
interface ApiResponse<T> {
  status: string;
  status_code: number;
  data: T;
  message?: string;
}

interface DistrictDTOResponse {
  id: string;
  name: string;
  region_id: string;
}

interface SchoolDTOResponse {
  id: string;
  name: string;
  district_id: string;
}

interface TeamCreateDTO {
  team_name: string;
  responsible_name: string;
  responsible_number: string;
  event_id: string;
  school_id: string;
}

interface ParticipantCreateDTO {
  FIO: string;
  gender: number;
  class: number;
  school_id: string;
}

interface RegisterTeamRequestDTO {
  team_name: string;
  responsible_name: string;
  responsible_number: string;
  event_id: string;
  school_id: string;
  participants: ParticipantCreateDTO[];
}

// ------------------------------------
// Компонент регистрации команды
// ------------------------------------
export default function TeamInitialize() {
  const { id } = useParams<{ id: string }>();

  const [districts, setDistricts] = useState<DistrictDTOResponse[]>([]);
  const [schools, setSchools] = useState<SchoolDTOResponse[]>([]);
  const [selectedDistrict, setSelectedDistrict] = useState("");
  const [loadingDistricts, setLoadingDistricts] = useState(false);
  const [loadingSchools, setLoadingSchools] = useState(false);

  const [team, setTeam] = useState<TeamCreateDTO>({
    team_name: "",
    responsible_name: "",
    responsible_number: "",
    event_id: id || "",
    school_id: "",
  });

  const [participants, setParticipants] = useState<ParticipantCreateDTO[]>([]);
  const [showModal, setShowModal] = useState(false);
  const [editIndex, setEditIndex] = useState<number | null>(null);

  const [participant, setParticipant] = useState<ParticipantCreateDTO>({
    FIO: "",
    gender: 1,
    class: 0,
    school_id: "",
  });

  const [submitting, setSubmitting] = useState(false);
  const [notify, setNotify] = useState<{ variant: "success" | "danger"; text: string } | null>(null);

  // ------------------------------------
  // Загрузка муниципалитетов
  // ------------------------------------
  useEffect(() => {
    const loadDistricts = async () => {
      setLoadingDistricts(true);
      try {
        const res = await fetch(`${API_URL}/api/locations/districts`);
        const json: ApiResponse<DistrictDTOResponse[]> = await res.json();
        if (res.ok) {
          setDistricts(json.data);
        } else {
          throw new Error(json.message || "Ошибка при загрузке муниципалитетов");
        }
      } catch (err) {
        console.error(err);
        setNotify({ variant: "danger", text: "Не удалось загрузить муниципалитеты" });
      } finally {
        setLoadingDistricts(false);
      }
    };
    loadDistricts();
  }, []);

  // ------------------------------------
  // Загрузка школ по выбранному муниципалитету
  // ------------------------------------
  useEffect(() => {
    const loadSchools = async () => {
      if (!selectedDistrict) {
        setSchools([]);
        setTeam((prev) => ({ ...prev, school_id: "" }));
        return;
      }

      setLoadingSchools(true);
      try {
        const res = await fetch(`${API_URL}/api/locations/districts/${selectedDistrict}/schools`);
        const json: ApiResponse<SchoolDTOResponse[]> = await res.json();
        if (res.ok) {
          setSchools(json.data);
        } else {
          throw new Error(json.message || "Ошибка при загрузке школ");
        }
      } catch (err) {
        console.error(err);
        setNotify({ variant: "danger", text: "Не удалось загрузить школы" });
      } finally {
        setLoadingSchools(false);
      }
    };
    loadSchools();
  }, [selectedDistrict]);

  // ------------------------------------
  // Работа с участниками
  // ------------------------------------
  const handleAddOrEditParticipant = () => {
    if (!participant.FIO.trim()) return "Введите ФИО участника";
    if (!participant.class) return "Укажите класс участника";

    const data = { ...participant, school_id: team.school_id };

    if (editIndex !== null) {
      const updated = [...participants];
      updated[editIndex] = data;
      setParticipants(updated);
      setEditIndex(null);
    } else {
      setParticipants([...participants, data]);
    }

    setParticipant({ FIO: "", gender: 1, class: 0, school_id: "" });
    setShowModal(false);
  };

  const handleEdit = (index: number) => {
    setParticipant(participants[index]);
    setEditIndex(index);
    setShowModal(true);
  };

  const handleDelete = (index: number) => {
    setParticipants(participants.filter((_, i) => i !== index));
  };

  // ------------------------------------
  // Отправка команды
  // ------------------------------------
  const handleSubmit = async () => {
    if (!team.team_name.trim()) return "Введите название команды";
    if (!team.school_id) return "Выберите школу";
    if (!participants.length) return "Добавьте хотя бы одного участника";

    const dto: RegisterTeamRequestDTO = {
      ...team,
      participants,
    };

    try {
      setSubmitting(true);
      const res = await fetch(`${API_URL}/api/teams/register`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(dto),
      });

      const json = await res.json();
      if (res.ok) {
        setNotify({ variant: "success", text: "Команда успешно зарегистрирована!" });
        setTeam({ ...team, team_name: "", responsible_name: "", responsible_number: "", school_id: "" });
        setParticipants([]);
      } else {
        throw new Error(json.message || "Ошибка при регистрации");
      }
    } catch (err) {
      console.error(err);
      setNotify({ variant: "danger", text: "Ошибка при отправке данных" });
    } finally {
      setSubmitting(false);
    }
  };

  // ------------------------------------
  // JSX
  // ------------------------------------
  return (
    <Container className="">
      <h2 className="text-center mb-4">Регистрация команды</h2>

      {notify && <Alert variant={notify.variant}>{notify.text}</Alert>}

      <Card className="mb-4 shadow-sm">
        <Card.Body>
          <h5 className="mb-3">Информация о команде</h5>
          <Form>
            <Form.Group className="mb-3">
              <Form.Label>Название команды *</Form.Label>
              <Form.Control
                value={team.team_name}
                onChange={(e) => setTeam({ ...team, team_name: e.target.value })}
                placeholder="Например: Лидеры будущего"
              />
            </Form.Group>

            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>ФИО ответственного</Form.Label>
                  <Form.Control
                    value={team.responsible_name}
                    onChange={(e) => setTeam({ ...team, responsible_name: e.target.value })}
                  />
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Телефон ответственного</Form.Label>
                  <Form.Control
                    value={team.responsible_number}
                    onChange={(e) => setTeam({ ...team, responsible_number: e.target.value })}
                    placeholder="+7..."
                  />
                </Form.Group>
              </Col>
            </Row>

            <Form.Group className="mb-3">
              <Form.Label>Муниципалитет *</Form.Label>
              <Select
                isDisabled={loadingDistricts}
                isLoading={loadingDistricts}
                placeholder="Выберите муниципалитет..."
                options={districts.map((d) => ({ value: d.id, label: d.name }))}
                value={districts
                  .filter((d) => d.id === selectedDistrict)
                  .map((d) => ({ value: d.id, label: d.name }))[0] || null}
                onChange={(opt) => {
                  setSelectedDistrict(opt ? opt.value : "");
                  setTeam((prev) => ({ ...prev, school_id: "" }));
                }}
                styles={customSelectStyles}
              />
            </Form.Group>

            <Form.Group className="mb-3">
              <Form.Label>Школа *</Form.Label>
              <Select
                isDisabled={!selectedDistrict || loadingSchools}
                isLoading={loadingSchools}
                placeholder={
                  loadingSchools
                    ? "Загрузка школ..."
                    : !selectedDistrict
                      ? "Сначала выберите муниципалитет"
                      : "Выберите школу..."
                }
                options={schools.map((s) => ({ value: s.id, label: s.name }))}
                value={schools
                  .filter((s) => s.id === team.school_id)
                  .map((s) => ({ value: s.id, label: s.name }))[0] || null}
                onChange={(opt) => setTeam({ ...team, school_id: opt ? opt.value : "" })}
               styles={customSelectStyles}
              />
            </Form.Group>

          </Form>
        </Card.Body>
      </Card>

      {/* Участники */}
      <Card className="mb-4 shadow-sm">
        <Card.Body>
          <div className="d-flex justify-content-between align-items-center mb-3">
            <h5>Участники</h5>
            <Button onClick={() => setShowModal(true)} disabled={!team.school_id}>
              + Добавить участника
            </Button>
          </div>

          <Table striped bordered hover responsive>
            <thead>
              <tr>
                <th>ФИО</th>
                <th>Пол</th>
                <th>Класс</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {participants.length > 0 ? (
                participants.map((p, i) => (
                  <tr key={i}>
                    <td>{p.FIO}</td>
                    <td>{p.gender === 1 ? "Мужской" : "Женский"}</td>
                    <td>{p.class}</td>
                    <td>
                      <Button size="sm" variant="outline-primary" className="me-2" onClick={() => handleEdit(i)}>
                        Редактировать
                      </Button>
                      <Button size="sm" variant="outline-danger" onClick={() => handleDelete(i)}>
                        Удалить
                      </Button>
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={4} className="text-center">Нет участников</td>
                </tr>
              )}
            </tbody>
          </Table>
        </Card.Body>
      </Card>

      <div className="text-center">
        <Button
          size="lg"
          onClick={handleSubmit}
          disabled={submitting || !team.school_id || !participants.length}
        >
          {submitting ? <Spinner animation="border" size="sm" /> : "Отправить заявку"}
        </Button>
      </div>

      {/* Модальное окно участника */}
      <Modal show={showModal} onHide={() => setShowModal(false)} centered>
        <Modal.Header closeButton>
          <Modal.Title>{editIndex !== null ? "Редактировать участника" : "Добавить участника"}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <Form>
            <Form.Group className="mb-3">
              <Form.Label>ФИО *</Form.Label>
              <Form.Control
                value={participant.FIO}
                onChange={(e) => setParticipant({ ...participant, FIO: e.target.value })}
                placeholder="Иванов Иван Иванович"
              />
            </Form.Group>

            <Form.Group className="mb-3">
              <Form.Label>Пол *</Form.Label>
              <Form.Select
                value={participant.gender}
                onChange={(e) => setParticipant({ ...participant, gender: parseInt(e.target.value) })}
              >
                <option value={2}>Женский</option>
                <option value={1}>Мужской</option>
              </Form.Select>
            </Form.Group>

            <Form.Group className="mb-3">
              <Form.Label>Класс *</Form.Label>
              <Form.Control
                type="number"
                min={1}
                max={11}
                value={participant.class}
                onChange={(e) => {
                  let value = parseInt(e.target.value) || 0;
                  if (value > 11) value = 11;
                  if (value < 1) value = 1;
                  setParticipant({ ...participant, class: value });
                }}
              />
            </Form.Group>
          </Form>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => setShowModal(false)}>
            Отмена
          </Button>
          <Button variant="primary" onClick={handleAddOrEditParticipant}>
            {editIndex !== null ? "Сохранить" : "Добавить"}
          </Button>
        </Modal.Footer>
      </Modal>
    </Container>
  );

  
}


const customSelectStyles = {
  control: (base: any) => ({
    ...base,
    borderRadius: "8px",
    backgroundColor: "#fff", // фон всегда белый
    color: "#000",           // текст всегда черный
    borderColor: "#ced4da",
    boxShadow: "none",
    "&:hover": {
      borderColor: "#86b7fe",
    },
  }),
  menu: (base: any) => ({
    ...base,
    backgroundColor: "#fff",
    color: "#000",
    zIndex: 9999,
  }),
  singleValue: (base: any) => ({
    ...base,
    color: "#000",
  }),
  input: (base: any) => ({
    ...base,
    color: "#000",
  }),
  placeholder: (base: any) => ({
    ...base,
    color: "#666",
  }),
  option: (base: any, state: any) => ({
    ...base,
    backgroundColor: state.isFocused ? "#e9ecef" : "#fff",
    color: "#000",
    cursor: "pointer",
  }),
};
