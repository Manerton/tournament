

const AdminPanel: React.FC = () => {
  return (
    <>
      <div
        className="d-flex flex-column justify-content-center align-items-center h-100 text-center p-4"
        style={{ minHeight: '70dvh' }}
      >
        <h1 className="display-4 fw-bold text-center">
          Администрирование<br />
          <hr />
          Турниры РШТ <br />
          <span
            style={{
              background: 'linear-gradient(to right, #1494D4, #70FF99)',
              WebkitBackgroundClip: 'text',
              WebkitTextFillColor: 'transparent',
              display: 'inline-block'
            }}
          >
            в Астраханской области!
          </span>
        </h1>

      </div>
    </>
  );
};

export default AdminPanel;