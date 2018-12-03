package msglib.proto.types;

public class MSet {
    public int ElementType = MType.Null;
    public int Count = 0;

    public MSet() {
    }

    public MSet(int elemtype, int count) {
        this.ElementType = elemtype;
        this.Count = count;
    }
}
