import { useContext } from 'react';
import { Button } from 'react-bootstrap';
import { Moon, Sun } from 'react-bootstrap-icons';
import { ThemeContext } from './ThemeContext'; // скорректируй путь при необходимости

const ThemeToggleButton = () => {
  const { theme, toggleTheme } = useContext(ThemeContext);

  return (
   <Button
        onClick={toggleTheme}
        variant={theme === 'light' ? 'dark' : 'light'}
        className="d-flex justify-content-center align-items-center ms-auto p-1"
        style={{
            width: '36px',
            height: '36px',
            borderRadius: '50%',
        }}
        aria-label={theme === 'light' ? 'Switch to Dark Mode' : 'Switch to Light Mode'}
        >
        {theme === 'light' ? <Moon size={20} /> : <Sun size={20} />}
    </Button>

  );
};

export default ThemeToggleButton;
