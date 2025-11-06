import { useState } from "react";
import { Button, Form, Alert, Spinner } from "react-bootstrap";
import type { SubmissionDTOResponse } from "../../types/submissionTypes";
import { axiosGetSubmissionByCode } from "../../requests/submissionRequest";

interface SidebarProps {
  eventId: string;
}

const SubmissionSidebar: React.FC<SidebarProps> = ({ eventId }) => {
  const [isOpen, setIsOpen] = useState(false);
  const [code, setCode] = useState("");
  const [result, setResult] = useState<SubmissionDTOResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!code.trim()) return;

    setLoading(true);
    setError("");
    setResult(null);

    try {
      const data = await axiosGetSubmissionByCode(eventId, code.trim());
      setResult(data);
      console.log("data", data)
    } catch (err: any) {
      setError(err.message || "Ошибка при поиске решения");
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      {/* Кнопка открытия */}
      {!isOpen && (
        <Button
          variant="primary"
          className="position-fixed top-50 end-0 translate-middle-y shadow"
          style={{ borderRadius: "8px 0 0 8px", zIndex: 1051 }}
          onClick={() => setIsOpen(true)}
        >
          Поиск
        </Button>
      )}

      {/* Сайдбар */}
      <div
        className={`position-fixed top-0 end-0 bg-light border-start shadow p-3 transition-all`}
        style={{
          width: "320px",
          height: "100vh",
          overflowY: "auto",
          transform: isOpen ? "translateX(0)" : "translateX(100%)",
          transition: "transform 0.3s ease-in-out",
          zIndex: 1050,
        }}
      >
        <div className="d-flex justify-content-between align-items-center mb-3">
          <h5 className="fw-bold mb-0">Поиск по коду</h5>
          <Button variant="outline-danger" size="sm" onClick={() => setIsOpen(false)}>
            X
          </Button>
        </div>

        <Form onSubmit={handleSearch}>
          <Form.Group className="mb-3">
            <Form.Control
              type="text"
              placeholder="Введите код задания..."
              value={code}
              onChange={(e) => setCode(e.target.value.toUpperCase())}
            />
          </Form.Group>
          <Button
            type="submit"
            variant="primary"
            className="w-100"
            disabled={loading}
          >
            {loading ? <Spinner size="sm" animation="border" /> : "Найти"}
          </Button>
        </Form>

        <div className="mt-4">
          {error && <Alert variant="danger">{error}</Alert>}
          {result && (
            <Alert variant="success" className="white-space-pre-wrap">
              <div>
                <strong>Задание:</strong>
                <div
                  className="bg-white border rounded p-3 mt-2"
                  style={{
                    whiteSpace: "pre-wrap",
                    wordBreak: "break-word",
                    fontSize: "1.1rem",
                    lineHeight: "1.5",
                    overflowX: "hidden",
                  }}
                >
                  {result.answer.replace(/;/g, "\n")}
                </div>

              </div>
            </Alert>
          )}
        </div>
      </div>
    </>
  );
};

export default SubmissionSidebar;
