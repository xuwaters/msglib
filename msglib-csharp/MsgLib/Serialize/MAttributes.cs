using System;
using System.Collections.Generic;
using System.Text;

namespace MsgLib
{
    [AttributeUsage(AttributeTargets.Class)]
    internal class MStructAttribute : Attribute
    {
    }

    [AttributeUsage(AttributeTargets.Property)]
	internal class MFieldAttribute : Attribute
    {
		internal int ID { get; set; }
    }

    [AttributeUsage(AttributeTargets.Property)]
	internal class MListAttribute : Attribute
    {
		internal Type ElemType { get; set; }
    }

    [AttributeUsage(AttributeTargets.Property)]
	internal class MMapAttribute : Attribute
    {
		internal Type KeyType { get; set; }
		internal Type ValueType { get; set; }
    }
}
