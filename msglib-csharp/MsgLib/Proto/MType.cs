using System;
using System.Collections.Generic;

namespace MsgLib
{
    internal enum MType : byte
    {
        Null = 1,
        Bool = 2,
        Byte = 3,
        I16 = 4,
        I32 = 5,
        I64 = 6,
        Float = 7,
        Double = 8,
        Binary = 9,
        String = 10,
        Struct = 11,
        Map = 12,
        List = 13,
        Set = 14,
    }
    
    internal enum MActionType : byte
    {
        Request = 1,
        Response = 2,
    }

    internal struct MList
    {
        internal MType ElementType { get; set; }
        internal int Count { get; set; }
        internal MList(MType elemtype, int count)
            : this()
        {
            this.ElementType = elemtype;
            this.Count = count;
        }
    }

    internal struct MSet
    {
        internal MType ElementType { get; set; }
        internal int Count { get; set; }
        internal MSet(MType elemtype, int count)
            : this()
        {
            this.ElementType = elemtype;
            this.Count = count;
        }
    }

    internal struct MMap
    {
        internal MType KeyType { get; set; }
        internal MType ValueType { get; set; }
        internal int Count { get; set; }
        internal MMap(MType keytype, MType valuetype, int count)
            : this()
        {
            this.KeyType = keytype;
            this.ValueType = valuetype;
            this.Count = count;
        }
    }

    internal struct MStruct
    {
        internal string Name { get; set; }
        internal MStruct(string name)
            : this()
        {
            this.Name = name;
        }
    }

    internal struct MField
    {
        internal string Name { get; set; }
        internal MType Type { get; set; }
        internal int ID { get; set; }
        internal MField(string name, MType type, int id)
            : this()
        {
            this.Name = name;
            this.Type = type;
            this.ID = id;
        }
    }

    internal struct MAction
    {
        internal string Name { get; set; }
        internal MActionType Type { get; set; }
        internal int ModuleID { get; set; }
        internal int ActionID { get; set; }
        internal int SeqID { get; set; }
        internal MAction(int moduleid, int actionid, string name, MActionType type, int seqid)
            : this()
        {
            this.ModuleID = moduleid;
            this.ActionID = actionid;
            this.Name = name;
            this.Type = type;
            this.SeqID = seqid;
        }
    }
}
