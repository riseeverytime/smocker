import { EditOutlined, SaveOutlined } from "@ant-design/icons";
import { Button, Card, Drawer, Form, PageHeader, Row } from "antd";
import * as React from "react";
import { connect } from "react-redux";
import { Dispatch } from "redux";
import { saveSvgAsPng } from "save-svg-as-png";
import { Actions, actions } from "~modules/actions";
import { AppState } from "~modules/reducers";
import Code from "./Code";
import { Mermaid } from "./Mermaid";
import "./Visualize.scss";

const EditGraph = ({
  display,
  value,
  onChange,
  onClose,
}: {
  display: boolean;
  value: string;
  onChange: (value: string) => void;
  onClose: () => void;
}) => {
  return (
    <Drawer
      title="Edit"
      placement="right"
      className="drawer"
      closable={true}
      onClose={onClose}
      visible={display}
      width="50vw"
      getContainer={false}
    >
      <Form className="form">
        <Code value={value} onBeforeChange={onChange} />
      </Form>
    </Drawer>
  );
};

interface Props {
  sessionID: string;
  graph: string;
  visualize: (sessionID: string) => void;
}

const Visualize = ({ sessionID, graph, visualize }: Props) => {
  const [diagram, setDiagram] = React.useState("");
  const [editGraph, setEditGraph] = React.useState(false);

  React.useEffect(() => {
    setDiagram(graph);
  }, [graph]);

  React.useEffect(() => {
    visualize(sessionID);
  }, [sessionID]);

  const handleEditGraph = () => setEditGraph(true);
  const handleChangeGraph = (diag: string) => setDiagram(diag);
  const handleCloseEditGraph = () => setEditGraph(false);
  const handleSaveGraph = () => {
    saveSvgAsPng(document.getElementById("diagram"), "diagram.png", {
      backgroundColor: "white",
      scale: 2,
      encoderOptions: 1,
      ignoreMouse: true,
      ignoreAnimation: true,
    });
  };
  return (
    <div className="visualize">
      <PageHeader
        title={"Diagram of calls"}
        extra={
          <div className="action buttons">
            <Button
              type="primary"
              icon={<EditOutlined />}
              onClick={handleEditGraph}
            >
              Edit
            </Button>
            <Button icon={<SaveOutlined />} onClick={handleSaveGraph}>
              Save
            </Button>
          </div>
        }
      >
        <p>This is a graphical representation of call history.</p>
        <Row justify="center" align="middle" className="container">
          {diagram && (
            <Card className={"card"}>
              <Mermaid name="diagram" chart={diagram} />
            </Card>
          )}
        </Row>
      </PageHeader>
      {editGraph && (
        <EditGraph
          display={editGraph}
          value={diagram}
          onClose={handleCloseEditGraph}
          onChange={handleChangeGraph}
        />
      )}
    </div>
  );
};

export default connect(
  (state: AppState) => {
    const { sessions, history } = state;
    return {
      sessionID: sessions.selected,
      graph: history.graph,
    };
  },
  (dispatch: Dispatch<Actions>) => ({
    visualize: (sessionID: string) =>
      dispatch(actions.visualizeHistory.request(sessionID)),
  })
)(Visualize);
