package msglib.proto.types;

public class MField {
    public String Name;
    public int Type;
    public int ID;

    public MField() {
    }

    public MField(String name, int type, int id) {
        this.Name = name;
        this.Type = type;
        this.ID = id;
    }
}
