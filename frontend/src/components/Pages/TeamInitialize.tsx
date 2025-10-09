import React, { useState, useEffect } from "react";
import { Container, Row, Col, Form, Button, Card, Table, Modal } from "react-bootstrap";

// Интерфейсы для данных
interface Municipality {
  id: string;
  name: string;
}

interface School {
  id: string;
  name: string;
  municipality_id: string;
}

interface ParticipantName {
  lastName: string;
  firstName: string;
  middleName: string;
}

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
  event_id?: string;
}

// Мок-данные муниципалитетов (будет заменено на запрос)
const mockMunicipalities: Municipality[] = [
  { id: "1", name: "Муниципалитет №1" },
  { id: "2", name: "Муниципалитет №2" },
  { id: "3", name: "Муниципалитет №3" },
];

// Мок-функция для загрузки школ (будет заменена на реальный запрос)
const fetchSchoolsByMunicipality = async (municipalityId: string): Promise<School[]> => {
  // Имитация задержки запроса
  await new Promise(resolve => setTimeout(resolve, 100));
  
  const mockSchools: Record<string, School[]> = {
    "1": [
      { id: "11111111-1111-1111-1111-111111111111", name: "Школа №1", municipality_id: "1" },
      { id: "11111111-1111-1111-1111-111111111112", name: "Школа №2", municipality_id: "1" },
    ],
    "2": [
      { id: "22222222-2222-2222-2222-222222222222", name: "Лицей №1", municipality_id: "2" },
      { id: "22222222-2222-2222-2222-222222222223", name: "Гимназия №1", municipality_id: "2" },
    ],
    "3": [
      { id: "33333333-3333-3333-3333-333333333333", name: "Школа №3", municipality_id: "3" },
    ],
  };
  
  return mockSchools[municipalityId] || [];
};

export default function RegisterTeamPage() {
  const [team, setTeam] = useState<Team>({
    team_name: "",
    responsible_name: "",
    responsible_number: "",
    school_id: "",
    event_id: "",
  });

  const [municipalities, setMunicipalities] = useState<Municipality[]>([]);
  const [selectedMunicipality, setSelectedMunicipality] = useState<string>("");
  const [schools, setSchools] = useState<School[]>([]);
  const [loadingSchools, setLoadingSchools] = useState<boolean>(false);

  const [participants, setParticipants] = useState<Participant[]>([]);
  const [showModal, setShowModal] = useState(false);
  const [currentParticipantName, setCurrentParticipantName] = useState<ParticipantName>({
    lastName: "",
    firstName: "",
    middleName: "",
  });
  const [currentParticipantDetails, setCurrentParticipantDetails] = useState({
    gender: 0,
    class: 0,
  });
  const [editIndex, setEditIndex] = useState<number | null>(null);

  // Загрузка муниципалитетов при монтировании компонента
  useEffect(() => {
    const loadMunicipalities = async () => {
      // Здесь будет реальный запрос к API
      // const response = await fetch('/api/municipalities');
      // const data = await response.json();
      setMunicipalities(mockMunicipalities);
    };

    loadMunicipalities();
  }, []);

  // Загрузка школ при изменении выбранного муниципалитета
  useEffect(() => {
    const loadSchools = async () => {
      if (!selectedMunicipality) {
        setSchools([]);
        setTeam(prev => ({ ...prev, school_id: "" }));
        return;
      }

      setLoadingSchools(true);
      try {
        const schoolsData = await fetchSchoolsByMunicipality(selectedMunicipality);
        setSchools(schoolsData);
        
        // Сбрасываем выбранную школу при смене муниципалитета
        setTeam(prev => ({ ...prev, school_id: "" }));
      } catch (error) {
        console.error("Ошибка при загрузке школ:", error);
        setSchools([]);
      } finally {
        setLoadingSchools(false);
      }
    };

    loadSchools();
  }, [selectedMunicipality]);

  const combineFIO = (name: ParticipantName): string => {
    return `${name.lastName} ${name.firstName} ${name.middleName}`.trim();
  };

  const splitFIO = (fio: string): ParticipantName => {
    const parts = fio.split(' ');
    return {
      lastName: parts[0] || "",
      firstName: parts[1] || "",
      middleName: parts[2] || "",
    };
  };

  const handleMunicipalityChange = (municipalityId: string) => {
    setSelectedMunicipality(municipalityId);
  };

  const handleAddParticipant = () => {
    if (!team.school_id) {
      alert("Сначала выберите школу команды!");
      return;
    }

    if (!currentParticipantName.lastName.trim() || !currentParticipantName.firstName.trim()) {
      alert("Заполните фамилию и имя участника!");
      return;
    }

    const participant: Participant = {
      fio: combineFIO(currentParticipantName),
      gender: currentParticipantDetails.gender,
      class: currentParticipantDetails.class,
      school_id: team.school_id,
    };

    if (editIndex !== null) {
      const updated = [...participants];
      updated[editIndex] = participant;
      setParticipants(updated);
      setEditIndex(null);
    } else {
      setParticipants([...participants, participant]);
    }

    setCurrentParticipantName({ lastName: "", firstName: "", middleName: "" });
    setCurrentParticipantDetails({ gender: 0, class: 0 });
    setShowModal(false);
  };

  const handleEditParticipant = (index: number) => {
    const participant = participants[index];
    setCurrentParticipantName(splitFIO(participant.fio));
    setCurrentParticipantDetails({
      gender: participant.gender,
      class: participant.class,
    });
    setEditIndex(index);
    setShowModal(true);
  };

  const handleDeleteParticipant = (index: number) => {
    setParticipants(participants.filter((_, i) => i !== index));
  };

  const handleSubmit = async () => {
    if (!team.school_id) {
      alert("Выберите школу!");
      return;
    }

    if (participants.length === 0) {
      alert("Добавьте хотя бы одного участника!");
      return;
    }

    const dto = {
      ...team,
      participants,
    };

    console.log("Отправка:", dto);

    try {
      await fetch("/api/team/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(dto),
      });
      alert("Заявка успешно отправлена!");
    } catch (error) {
      console.error("Ошибка при отправке:", error);
      alert("Произошла ошибка при отправке заявки");
    }
  };

  return (
    <Container className="my-4">
      <h2 className="text-center mb-4">Регистрация команды</h2>

      {/* --- Данные команды --- */}
      <Card className="mb-4 shadow-sm">
        <Card.Body>
          <h5 className="mb-3">Информация о команде</h5>
          <Form>
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

            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>ФИО ответственного</Form.Label>
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

            {/* Выбор муниципалитета */}
            <Form.Group className="mb-3">
              <Form.Label>Муниципалитет *</Form.Label>
              <Form.Select
                value={selectedMunicipality}
                onChange={(e) => handleMunicipalityChange(e.target.value)}
              >
                <option value="">Выберите муниципалитет</option>
                {municipalities.map((municipality) => (
                  <option key={municipality.id} value={municipality.id}>
                    {municipality.name}
                  </option>
                ))}
              </Form.Select>
            </Form.Group>

            {/* Выбор школы */}
            <Form.Group className="mb-3">
              <Form.Label>Школа *</Form.Label>
              <Form.Select
                value={team.school_id}
                onChange={(e) =>
                  setTeam({ ...team, school_id: e.target.value })
                }
                disabled={!selectedMunicipality || loadingSchools}
              >
                <option value="">
                  {loadingSchools 
                    ? "Загрузка школ..." 
                    : !selectedMunicipality 
                    ? "Сначала выберите муниципалитет" 
                    : "Выберите школу"}
                </option>
                {schools.map((school) => (
                  <option key={school.id} value={school.id}>
                    {school.name}
                  </option>
                ))}
              </Form.Select>
              {loadingSchools && (
                <Form.Text className="text-muted">
                  Загружаем список школ...
                </Form.Text>
              )}
            </Form.Group>
          </Form>
        </Card.Body>
      </Card>

      {/* --- Участники --- */}
      <Card className="mb-4 shadow-sm">
        <Card.Body>
          <div className="d-flex justify-content-between align-items-center mb-3">
            <h5>Участники</h5>
            <Button 
              onClick={() => setShowModal(true)}
              disabled={!team.school_id}
            >
              + Добавить
            </Button>
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
        <Button 
          size="lg" 
          onClick={handleSubmit}
          disabled={!team.school_id || participants.length === 0}
        >
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
            <Row>
              <Col md={4}>
                <Form.Group className="mb-3">
                  <Form.Label>Фамилия *</Form.Label>
                  <Form.Control
                    type="text"
                    value={currentParticipantName.lastName}
                    onChange={(e) =>
                      setCurrentParticipantName({ 
                        ...currentParticipantName, 
                        lastName: e.target.value 
                      })
                    }
                    placeholder="Иванов"
                  />
                </Form.Group>
              </Col>
              <Col md={4}>
                <Form.Group className="mb-3">
                  <Form.Label>Имя *</Form.Label>
                  <Form.Control
                    type="text"
                    value={currentParticipantName.firstName}
                    onChange={(e) =>
                      setCurrentParticipantName({ 
                        ...currentParticipantName, 
                        firstName: e.target.value 
                      })
                    }
                    placeholder="Иван"
                  />
                </Form.Group>
              </Col>
              <Col md={4}>
                <Form.Group className="mb-3">
                  <Form.Label>Отчество</Form.Label>
                  <Form.Control
                    type="text"
                    value={currentParticipantName.middleName}
                    onChange={(e) =>
                      setCurrentParticipantName({ 
                        ...currentParticipantName, 
                        middleName: e.target.value 
                      })
                    }
                    placeholder="Иванович"
                  />
                </Form.Group>
              </Col>
            </Row>

            <Form.Group className="mb-3">
              <Form.Label>Пол</Form.Label>
              <Form.Select
                value={currentParticipantDetails.gender}
                onChange={(e) =>
                  setCurrentParticipantDetails({ 
                    ...currentParticipantDetails, 
                    gender: parseInt(e.target.value) 
                  })
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
                value={currentParticipantDetails.class}
                onChange={(e) =>
                  setCurrentParticipantDetails({ 
                    ...currentParticipantDetails, 
                    class: parseInt(e.target.value) || 0 
                  })
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