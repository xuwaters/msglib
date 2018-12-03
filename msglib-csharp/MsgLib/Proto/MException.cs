using System;

namespace MsgLib
{
    internal class MException : Exception
    {
		internal MException()
		{
		}

        internal MException(string message)
			: base(message)
		{
		}
    }
}
