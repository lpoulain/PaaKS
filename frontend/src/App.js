import { useCookies } from "react-cookie";
import Frontend from './index';
import logo from './logo.svg';
import './App.css';

function App() {
  const [cookies, setCookie] = useCookies(["member"]);

  return (
    <Frontend token={cookies["token"]} setCookie={setCookie}></Frontend>
  );
}

export default App;
