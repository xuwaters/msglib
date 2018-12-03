package msglib.proto.types;

public class MList {
    public int ElementType = MType.Null;
    public int Count = 0;

    public MList() {
    }

    public MList(int elemtype, int count) {
        this.ElementType = elemtype;
        this.Count = count;
    }
}
